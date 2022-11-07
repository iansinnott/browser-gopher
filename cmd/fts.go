/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/tui"
	"github.com/spf13/cobra"
)

// dbPathCmd represents the dbPath command
var ftsCmd = &cobra.Command{
	Use:   "fts",
	Short: "Full-text search",
	Long: `
Search the full text of web pages. Note that this requires web pgaes to already
have been indexed.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		initialQuery := ""

		if len(args) > 0 {
			initialQuery = args[0]
		}

		// result, err := provider.GetFullTextUrls(query)
		// if err != nil {
		// 	fmt.Println("error searching", err)
		// 	os.Exit(1)
		// }

		dataProvider := search.NewSqlSearchProvider(cmd.Context(), config.Config)
		searchProvider := search.NewBleveSearchProvider(cmd.Context(), config.Config)

		fmt.Println("todo")
		p, err := tui.GetSearchProgram(cmd.Context(), initialQuery, dataProvider, searchProvider)
		if err != nil {
			fmt.Println("could not get search program:", err)
			os.Exit(1)
		}

		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ftsCmd)
}
