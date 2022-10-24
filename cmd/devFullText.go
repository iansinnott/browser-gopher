package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
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
		fmt.Println(urlMd5, targetUrl)

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
			ua, _ := GetUserAgent()

			collector := colly.NewCollector(
				colly.UserAgent(ua),
				colly.CacheDir(cacheDir), // without cachedir colly will re-request every site (which may be what you want, just note)
				colly.MaxDepth(1),        // 0 means unlimited. not sure how this actually works since I thought it does NOT spider by default
				colly.AllowedDomains(hostname),
				colly.Async(false),
				colly.IgnoreRobotsTxt(),
			)

			collector.OnRequest(func(r *colly.Request) {
				fmt.Println("Fetching", r.URL)
			})

			collector.OnResponse(func(r *colly.Response) {
				fmt.Println("Fetched", len(r.Body), "bytes")
				html = r.Body
			})

			collector.OnError(func(r *colly.Response, err error) {
				fmt.Fprintf(os.Stderr, "error: %v %s\n", r.StatusCode, err)
				os.Exit(1)
			})

			collector.Visit(targetUrl)

			// unneeded since we're not async, but if you decide to go async you may need this
			collector.Wait()
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
