package cmd

import (
	"fmt"
	"os"

	bs "github.com/blevesearch/bleve/v2/search"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/search"
	"github.com/samber/lo"
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
			fmt.Println(hit.Fields["url"])
			fmt.Printf("terms: %+v\n", hit.Locations)
		}

		urls, err := searchProvider.SearchUrls("github")
		if err != nil {
			fmt.Println("search error", err)
			os.Exit(1)
		}

		for _, url := range urls.Urls {
			hit, ok := lo.Find(urls.Meta.Hits, func(x *bs.DocumentMatch) bool {
				return x.ID == url.UrlMd5
			})
			if ok {

				for k, locations := range hit.Locations {
					var s string
					switch k {
					case "title":
						s = *url.Title
					case "url":
						s = url.Url
					default:
					}

					for _, locs := range locations {
						for _, loc := range locs {
							s = highlightLocation(loc, s)
						}
					}

					fmt.Println(s)
				}

			}
		}
	},
}

func init() {
	devCmd.AddCommand(devBleveCmd)
}
