package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/iansinnott/browser-gopher/pkg/fulltext"
	"github.com/iansinnott/browser-gopher/pkg/logging"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/spf13/cobra"
	stripmd "github.com/writeas/go-strip-markdown"
)

// get user agent returns a valid user agent for use in scraping. in the future
// the idea is to have it generated at runtime, either by reading from local
// data or calling a remote api. thus the error return value.
func GetUserAgent() (string, error) {
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36", nil
}

var devFullTextCmd = &cobra.Command{
	Use:   "full-text",
	Short: "Get the full text of URLs",
	Long: `
Get the full text of a URL or stdin. This is used for dev in order to
easily check the FTS output of a given site. FTS processing is done
automatically for URLs.

Example:

	browser-gopher dev full-text 'https://example.com'
	curl 'https://example.com' | browser-gopher dev full-text -

	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No url provided")
			cmd.Help()
			os.Exit(1)
		}

		targetUrl := args[0]
		urlMd5 := util.HashMd5String(targetUrl)
		logging.Debug().Println("processing", urlMd5, targetUrl)

		var html []byte
		var err error
		var hostname string
		var pathname string

		cacheDir := filepath.Join("tmp", "scrape_cache")
		err = os.MkdirAll(cacheDir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %s\n", err)
			os.Exit(1)
		}

		if targetUrl == "-" {
			html, err = io.ReadAll(os.Stdin)
			hostname = "stdin"
			pathname = "-"
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
		} else {
			u, err := url.Parse(targetUrl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not parse: %s\n", err)
				os.Exit(1)
			}

			hostname = u.Hostname()
			pathname = strings.ReplaceAll(strings.Trim(u.Path, "/"), "/", "_")
			scraper := fulltext.NewScraper()
			htmls, err := scraper.ScrapeUrls([]string{targetUrl})

			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s ", err)
				os.Exit(1)
			}

			if len(htmls) != 1 {
				fmt.Fprintf(os.Stderr, "no html body found")
				os.Exit(1)
			}

			// @note the urls in the htmls map may not match the passed-in URLs. this is not a good API
			html = htmls[targetUrl].Body
		}

		outFile := fmt.Sprintf("%s_%s_%s", urlMd5, hostname, pathname)

		var outPath string
		outPath = filepath.Join("tmp", fmt.Sprintf("%s.html", outFile))
		err = os.WriteFile(outPath, html, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("wrote: " + outPath)

		converter := md.NewConverter(targetUrl, true, nil)
		bs, err := converter.ConvertBytes(html)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		outPath = filepath.Join("tmp", fmt.Sprintf("%s.md", outFile))
		err = os.WriteFile(outPath, bs, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "html2markdown: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("wrote: " + outPath)

		outPath = filepath.Join("tmp", fmt.Sprintf("%s.txt", outFile))
		err = os.WriteFile(outPath, []byte(stripmd.Strip(string(bs))), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "strip error: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("wrote: " + outPath)
	},
}

func init() {
	devCmd.AddCommand(devFullTextCmd)
}
