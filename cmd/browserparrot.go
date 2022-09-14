/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
)

// browserparrotCmd represents the browserparrot command
var browserparrotCmd = &cobra.Command{
	Use:   "browserparrot",
	Short: "Import URL history from a BrowserParrot database",
	Long: `If you have not previously used BrowserParrot this does not apply. This
command will import all URLs from BrowserParrot since you may already have
many URLs in there which are no longer present in the history databases of the
original browsers.

Using the command without any args will try the default location for the
BrowserParrot database, and should work in most cases.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, err := cmd.Flags().GetString("db-path")
		if err != nil {
			fmt.Println("could not parse --db-path:", err)
			os.Exit(1)
		}

		browserparrot := &extractors.BrowserParrotExtractor{
			HistoryDBPath: util.Expanduser(dbPath),
			Name:          "browserparrot",
		}
		err = PopulateAll(browserparrot)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(browserparrotCmd)
	browserparrotCmd.Flags().String("db-path", "~/.config/persistory/persistory.db", "The path to the database")
}
