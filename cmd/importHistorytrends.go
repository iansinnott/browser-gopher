package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/extractors"
	"github.com/iansinnott/browser-gopher/pkg/populate"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
)

var historytrendsCmd = &cobra.Command{
	Use:   "historytrends",
	Short: "Import from the History Trends Unlimited browser extension.",
	Long: `Using the command without any args will try the default location for
the BrowserParrot database, and should work in most cases.`,
	Run: func(cmd *cobra.Command, args []string) {
		searchPath, err := cmd.Flags().GetString("search-path")
		if err != nil {
			fmt.Println("could not parse --db-path:", err)
			os.Exit(1)
		}
		if searchPath == "" {
			fmt.Println("--search-path is required")
			os.Exit(1)
		}

		dbs, err := extractors.FindHistoryTrendsDBs(util.Expanduser(searchPath))
		if err != nil {
			fmt.Println("", err)
			os.Exit(1)
		}

		if len(dbs) == 0 {
			fmt.Println("History Trends Unlimited does not appear to be installed. Could not find it under the root path: ", searchPath)
			os.Exit(0)
		}

		for _, dbPath := range dbs {
			extractor := &extractors.HistoryTrendsExtractor{
				HistoryDBPath: util.Expanduser(dbPath),
				Name:          "historytrends",
			}
			fmt.Println("importing:", dbPath)
			err = populate.PopulateAll(extractor)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Println("Done.")
	},
}

func init() {
	importCmd.AddCommand(historytrendsCmd)
	historytrendsCmd.Flags().String("search-path", "~/Library/Application Support/Google/Chrome/", "The path to the database")
}
