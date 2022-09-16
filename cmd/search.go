package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Find URLs you've visited",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		searchProvider := search.NewSearchProvider(cmd.Context(), config.Config)
		result, err := searchProvider.SearchUrls(query)
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}
		for _, x := range util.ReverseSlice(result.Urls) {
			fmt.Printf("%v %s %sv\n", x.LastVisit.Format("2006-01-02"), *x.Title, x.Url)
		}
		fmt.Printf("Found %d results for \"%s\"\n", result.Count, query)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
