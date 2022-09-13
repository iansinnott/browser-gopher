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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := cmd.Flag("db-path").Value.String()
		browserparrot := &extractors.BrowserParrotExtractor{
			HistoryDBPath: util.Expanduser(dbPath),
			Name:          "browserparrot",
		}
		err := PopulateAll(browserparrot)
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
