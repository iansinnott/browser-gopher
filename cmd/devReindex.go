package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/populate"
	"github.com/spf13/cobra"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Reindex all URL records in the search index",
	Run: func(cmd *cobra.Command, args []string) {
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			fmt.Println("could not parse --limit:", err)
			os.Exit(1)
		}

		dbConn, err := persistence.InitDb(cmd.Context(), config.Config)
		if err != nil {
			fmt.Println("could not open our db", err)
			os.Exit(1)
		}

		err = os.RemoveAll(config.Config.SearchIndexPath)
		if err != nil {
			fmt.Println("could not remove search index", err)
			os.Exit(1)
		}

		fmt.Println("Reindexing everything...")
		t := time.Now()
		n, err := populate.ReindexWithLimit(cmd.Context(), dbConn, limit)
		if err != nil {
			fmt.Println("encountered an error building the search index", err)
			os.Exit(1)
		}
		fmt.Printf("Indexed %d records in %v\n", n, time.Since(t))
	},
}

func init() {
	reindexCmd.Flags().Int("limit", 0, "Limit the number of records to index")
	devCmd.AddCommand(reindexCmd)
}
