/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"

	ex "github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/populate"
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

		extractors, err := ex.BuildExtractorList()
		if err != nil {
			log.Println("error getting extractors", err)
			os.Exit(1)
		}

		if browserName != "" {
			for _, x := range extractors {
				if x.GetName() == browserName {
					err = populate.PopulateAll(x)
					if err != nil {
						log.Printf("Error with extractor: %+v\n", x)
					}
				}
			}
		} else {
			errs := []error{}

			// Without a browser name, populate everything
			for _, x := range extractors {
				e := populate.PopulateAll(x)
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
	populateCmd.Flags().StringP("browser", "b", "", "Specify the browser name you'd like to extract")
}
