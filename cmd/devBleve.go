package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/spf13/cobra"
)

var devBleveCmd = &cobra.Command{
	Use:   "bleve-search",
	Short: "Search bleve directly",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		if query == "" {
			fmt.Println("query is required")
			os.Exit(1)
		}

		searchProvider := search.NewBleveSearchProvider(cmd.Context(), config.Config)
		result, err := searchProvider.SearchBleve(query, "id", "url", "title", "description", "last_visit")
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		err = json.NewEncoder(os.Stdout).Encode(result)
		if err != nil {
			fmt.Println("json error", err)
			os.Exit(1)
		}

	},
}

func init() {
	devCmd.AddCommand(devBleveCmd)
}
