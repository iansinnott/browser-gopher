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
	Run: func(cmd *cobra.Command, args []string) {
		noInteractive, err := cmd.Flags().GetBool("no-interactive")
		if err != nil {
			fmt.Println("could not parse --no-interactive:", err)
			os.Exit(1)
		}

		searchProvider := search.NewSearchProvider(cmd.Context(), config.Config)

		if noInteractive {
			if len(args) < 1 {
				fmt.Println("No search query provided.")
				os.Exit(1)
				return
			}

			query := args[0]
			result, err := searchProvider.SearchUrls(query)
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
				return
			}

			for _, x := range util.ReverseSlice(result.Urls) {
				fmt.Printf("%v %s %sv\n", x.LastVisit.Format("2006-01-02"), *x.Title, x.Url)
			}

			fmt.Printf("Found %d results for \"%s\"\n", result.Count, query)
			os.Exit(0)
			return
		}

		fmt.Println("set up tui")
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
}
