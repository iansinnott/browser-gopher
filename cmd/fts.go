/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/logging"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/tui"
	"github.com/iansinnott/browser-gopher/pkg/util"
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

		dataProvider := search.NewSqlSearchProvider(cmd.Context(), config.Config)
		fullTextProvider := search.NewFullTextSearchProvider(cmd.Context(), config.Config)

		// check the no-interactive flag
		noInteractive, err := cmd.Flags().GetBool("no-interactive")
		if err != nil {
			fmt.Println("could not parse --no-interactive:", err)
			os.Exit(1)
		}

		if noInteractive {
			if len(args) < 1 {
				fmt.Println("No search query provided.")
				os.Exit(1)
			}

			result, err := fullTextProvider.SearchUrls(initialQuery)
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
			}

			for _, x := range util.ReverseSlice(result.Urls) {
				var displayTitle string
				if x.Title != nil {
					displayTitle = *x.Title
				} else {
					displayTitle = "<UNTITLED>"
				}
				fmt.Printf("%v %v\n", displayTitle, x.Url)
			}

			fmt.Printf("Found %d results for \"%s\"\n", result.Count, initialQuery)
			os.Exit(0)
			return
		}

		mapper := func(x tui.ListItem) list.Item {
			if x.Body == nil {
				logging.Debug().Printf("Mapping item %+v", x)
				x.ItemTitle = *x.Body
			}
			return list.Item(x)
		}

		p, err := tui.GetSearchProgram(cmd.Context(), initialQuery, dataProvider, fullTextProvider, &mapper)

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
	ftsCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
	rootCmd.AddCommand(ftsCmd)
}
