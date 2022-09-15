package populate

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

// inceptionTime is just an early time, assuming all observations will be after this time.
var inceptionTime time.Time = time.Unix(0, 0) // 1970-01-01

// PopulateAll populates all records from browsers, ignoring the last updated time
func PopulateAll(extractor types.Extractor) error {
	return PopulateSinceTime(extractor, inceptionTime)
}

func PopulateSinceTime(extractor types.Extractor, since time.Time) error {
	if since != inceptionTime {
		log.Println("["+extractor.GetName()+"] populating records since", since.String())
	}
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
		return PopulateSinceTime(extractor, since)
	}

	urls, err := extractor.GetAllUrlsSince(ctx, conn, since)
	if err != nil {
		return err
	}

	visits, err := extractor.GetAllVisitsSince(ctx, conn, since)
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
