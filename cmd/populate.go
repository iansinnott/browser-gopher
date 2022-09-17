/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	ex "github.com/iansinnott/browser-gopher/pkg/extractors"
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

		extractors, err := ex.BuildExtractorList()
		if err != nil {
			log.Println("error getting extractors", err)
			os.Exit(1)
		}

		errs := []error{}

		// Without a browser name, populate everything
		for _, x := range extractors {
			if browserName != "" && x.GetName() != browserName {
				continue
			}

			since := time.Unix(0, 0) // 1970-01-01
			if onlyLatest {
				dbConn, err := sql.Open("sqlite", config.Config.DBPath)
				if err != nil {
					fmt.Println("could not open our db", err)
					os.Exit(1)
				}

				latestTime, err := persistence.GetLatestTime(cmd.Context(), dbConn, x)
				if err != nil {
					fmt.Println("could not get latest time", err)
					os.Exit(1)
				}

				dbConn.Close()

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
				log.Println(e)
			}
			err = fmt.Errorf("one or more browsers failed")
		}

		if err != nil {
			fmt.Println("Encountered an error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(populateCmd)
	populateCmd.Flags().StringP("browser", "b", "", "Specify the browser name you'd like to extract")
	populateCmd.Flags().Bool("latest", false, "Only populate data that's newer than last import (Recommended, likely will be default in future version)")
}
