package cmd

import (
	"fmt"
	"os"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/iansinnott/browser-gopher/pkg/tui"
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

		dataProvider := search.NewSqlSearchProvider(cmd.Context(), config.Config)
		searchProvider := search.NewBleveSearchProvider(cmd.Context(), config.Config)
		initialQuery := ""

		if len(args) > 0 {
			initialQuery = args[0]
		}

		if noInteractive {
			if len(args) < 1 {
				fmt.Println("No search query provided.")
				os.Exit(1)
				return
			}

			result, err := searchProvider.SearchUrls(initialQuery)
			if err != nil {
				fmt.Println("search error", err)
				os.Exit(1)
				return
			}

			for _, x := range util.ReverseSlice(result.Urls) {
				fmt.Printf("%v %s %sv\n", x.LastVisit.Format("2006-01-02"), *x.Title, x.Url)
			}

			fmt.Printf("Found %d results for \"%s\"\n", result.Count, initialQuery)
			os.Exit(0)
			return
		}

		p, err := tui.GetSearchProgram(cmd.Context(), initialQuery, dataProvider, searchProvider, nil)
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
	searchCmd.Flags().Bool("no-interactive", false, "disable interactive terminal interface. useful for scripting")
	rootCmd.AddCommand(searchCmd)
}
