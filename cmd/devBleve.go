package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/spf13/cobra"
)

var devBleveCmd = &cobra.Command{
	Use:   "bleve-search",
	Short: "Search bleve directly",
	Run: func(cmd *cobra.Command, args []string) {
		searchProvider := search.NewBleveSearchProvider(cmd.Context(), config.Config)
		result, err := searchProvider.SearchBleve("github")
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		fmt.Println("count:", result.Total)
		for _, hit := range result.Hits {
			fmt.Printf("hit: %s\n", hit.ID)
		}

		urls, err := searchProvider.SearchUrls("github")
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		for _, url := range urls.Urls {
			fmt.Printf("url: %+v\n", url)
		}
	},
}

func init() {
	devCmd.AddCommand(devBleveCmd)
}
