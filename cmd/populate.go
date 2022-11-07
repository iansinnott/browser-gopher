/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	ex "github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/logging"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/populate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Populate URLs from all known sources",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		browserName, err := cmd.Flags().GetString("browser")
		if err != nil {
			fmt.Println("could not parse --browser:", err)
			os.Exit(1)
		}

		onlyLatest, err := cmd.Flags().GetBool("latest")
		if err != nil {
			fmt.Println("could not parse --latest:", err)
			os.Exit(1)
		}

		shouldBuildIndex, err := cmd.Flags().GetBool("build-index")
		if err != nil {
			fmt.Println("could not parse --build-index:", err)
			os.Exit(1)
		}

		shouldScrapeFulltext, err := cmd.Flags().GetBool("fulltext")
		if err != nil {
			fmt.Println("could not parse --fulltext:", err)
			os.Exit(1)
		}

		extractors, err := ex.BuildExtractorList()
		if err != nil {
			log.Println("error getting extractors", err)
			os.Exit(1)
		}

		dbConn, err := persistence.InitDb(cmd.Context(), config.Config)
		if err != nil {
			fmt.Println("could not open our db", err)
			os.Exit(1)
		}
		defer dbConn.Close()

		errs := []error{}

		// Without a browser name, populate everything
		for _, x := range extractors {
			if browserName != "" && x.GetName() != browserName {
				continue
			}

			since := time.Unix(0, 0) // 1970-01-01
			if onlyLatest {
				latestTime, err := persistence.GetLatestTime(cmd.Context(), dbConn, x)
				if err != nil {
					fmt.Println("could not get latest time", err)
					os.Exit(1)
				}

				since = *latestTime
			}

			var err error
			if onlyLatest {
				err = populate.PopulateSinceTime(x, since)
			} else {
				err = populate.PopulateAll(x)
			}
			if err != nil {
				errs = append(errs, errors.Wrap(err, x.GetName()+" error:"))
			}
		}

		if len(errs) > 0 {
			for _, e := range errs {
				logging.Warn().Println("browser failure:", e)
			}

			if len(errs) == len(extractors) {
				err = fmt.Errorf("all browsers failed to populate. exiting")
			}
		}

		if err != nil {
			logging.Error().Println("Encountered an error", err)
			os.Exit(1)
		}

		if shouldScrapeFulltext {
			var n int
			retries := 5
			t := time.Now()

			// @note It's not clear why sqlite is throwing busy errors. Concurrency is
			// used under the hood by colly but not directly in our code, so in theory
			// there should be only one goroutine accessing the database.
			// The retry loop is a workaround for episodic sqlite busy errors.
			for retries > 0 {
				n, err = populate.PopulateFulltext(cmd.Context(), dbConn)
				if err != nil {
					// if the error is sqlite_busy then retry once
					if strings.Contains(err.Error(), "database is locked") {
						fmt.Println("database is locked, retrying in 5 seconds")
						time.Sleep(5 * time.Second)
						retries--
						continue
					}

					logging.Error().Printf("could not populate fulltext: %v\n", err)
					os.Exit(1)
				}

				// if no error then break out
				break
			}

			log.Printf("Scraped %d pages in %v\n", n, time.Since(t))
		}

		if shouldBuildIndex {
			fmt.Println("Indexing results...")
			t := time.Now()
			n, err := populate.BuildIndex(cmd.Context(), dbConn)
			if err != nil {
				logging.Error().Printf("building the search index: %v\n", err)
				os.Exit(1)
			}
			log.Printf("Indexed %d records in %v\n", n, time.Since(t))
		}
	},
}

func init() {
	rootCmd.AddCommand(populateCmd)
	populateCmd.Flags().StringP("browser", "b", "", "Specify the browser name you'd like to extract")
	populateCmd.Flags().Bool("latest", false, "Only populate data that's newer than last import (Recommended, likely will be default in future version)")
	populateCmd.Flags().Bool("build-index", true, "Whether or not to build the search index. Required for search to work.")
	populateCmd.Flags().Bool("fulltext", false, "Whether or not to collect the full-text of each page in your browsing history and make it searchable.")
}
