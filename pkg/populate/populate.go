package populate

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

func PopulateAll(extractor types.Extractor) error {
	log.Println("["+extractor.GetName()+"] reading", extractor.GetDBPath())
	conn, err := sql.Open("sqlite", extractor.GetDBPath())
	ctx := context.TODO()

	if err != nil {
		log.Println("could not connect to db at", extractor.GetDBPath(), err)
		return err
	}
	defer conn.Close()

	// Handle the case where the database is in use, or return if the database cannot be read or copied.
	_, err = extractor.VerifyConnection(ctx, conn)
	if err != nil {
		if !strings.Contains(err.Error(), "SQLITE_BUSY") {
			log.Println("[err] Could read from DB", extractor.GetDBPath())
			return err
		}

		log.Println("[" + extractor.GetName() + "] database is locked. copying for read access.")

		tmpPath := filepath.Join(os.TempDir(), extractor.GetName()+"_backup.sqlite")

		err := util.CopyPath(extractor.GetDBPath(), tmpPath)
		if err != nil {
			fmt.Println("could not copy:", tmpPath)
			return err
		}
		// Remove interim file afterwards (otherwise these files eventually take up quite a bit of space)
		defer func() {
			if os.Remove(tmpPath) != nil {
				log.Println("could not remove tmp file:", tmpPath)
			}
		}()

		if extractor.GetDBPath() == tmpPath {
			return fmt.Errorf("recursive populate call detected. db tmp path must be different than initial db path")
		}

		// Update extractor to use the tmp path
		extractor.SetDBPath(tmpPath)

		// Retry with udpated db path
		return PopulateAll(extractor)
	}

	urls, err := extractor.GetAllUrls(ctx, conn)
	if err != nil {
		return err
	}

	visits, err := extractor.GetAllVisits(ctx, conn)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("["+extractor.GetName()+"]\turls", len(urls))
	log.Println("["+extractor.GetName()+"]\tvisits", len(visits))

	db, err := persistence.InitDB(ctx, config.Config)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, x := range urls {
		err := persistence.InsertURL(ctx, db, &x)
		if err != nil {
			log.Println("could not insert row", err)
		}
	}

	for _, x := range visits {
		if x.ExtractorName == "" {
			x.ExtractorName = extractor.GetName()
		}

		err := persistence.InsertVisit(ctx, db, &x)
		if err != nil {
			log.Println("could not insert row", err)
		}
	}

	return nil
}
