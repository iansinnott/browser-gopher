/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iansinnott/browser-gopher/pkg/config"
	ex "github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
)

// populateCmd represents the populate command
var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Populate URLs from all known sources",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		browserName := cmd.Flag("browser").Value.String()
		// browserName := flag.String("browser", "", "Specify which browser")
		// flag.Parse()

		extractors, err := ex.BuildExtractorList()
		if err != nil {
			log.Println("error getting extractors", err)
			os.Exit(1)
		}

		if browserName != "" {
			for _, x := range extractors {
				if x.GetName() == browserName {
					err = PopulateAll(x)
					if err != nil {
						log.Printf("Error with extractor: %+v\n", x)
					}
				}
			}
		} else {
			errs := []error{}

			// Without a browser name, populate everything
			for _, x := range extractors {
				e := PopulateAll(x)
				if e != nil {
					errs = append(errs, e)
				}
			}

			if len(errs) > 0 {
				for _, e := range errs {
					log.Println(e)
				}
				err = fmt.Errorf("one or more browsers failed")
			}
		}

		if err != nil {
			fmt.Println("Encountered an error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(populateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// populateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	populateCmd.Flags().StringP("browser", "b", "", "Specify the browser name you'd like to extract")
}

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

		// Update extractor to use the tmp path
		extractor.SetDBPath(tmpPath)

		defer func() {
			if os.Remove(tmpPath) != nil {
				log.Println("could not remove tmp file:", tmpPath)
			}
		}()

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

	log.Println("["+extractor.GetName()+"]\tfound urls", len(urls))
	log.Println("["+extractor.GetName()+"]\tfound visits", len(visits))

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
