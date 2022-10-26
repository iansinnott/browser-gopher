package fulltext

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/iansinnott/browser-gopher/pkg/logging"
)

type WebPage struct {
	Url        string
	Body       []byte
	Redirected bool
}

type Scraper struct {
	collector    *colly.Collector
	scrapedPages map[string]WebPage
	redirects    map[string]string
}

// get user agent returns a valid user agent for use in scraping. in the future
// the idea is to have it generated at runtime, either by reading from local
// data or calling a remote api. thus the error return value.
func GetUserAgent() (string, error) {
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36", nil
}

func NewScraper() *Scraper {
	ua, _ := GetUserAgent()
	collector := colly.NewCollector(
		colly.UserAgent(ua),
		// colly.CacheDir(cacheDir), // without cachedir colly will re-request every site (which may be what you want, just note)
		colly.MaxDepth(1), // 0 means unlimited. not sure how this actually works since I thought it does NOT spider by default
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)

	scraper := &Scraper{
		collector:    collector,
		scrapedPages: map[string]WebPage{},
		redirects:    map[string]string{},
	}

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Fetching", r.URL)
	})

	collector.OnResponseHeaders(func(r *colly.Response) {
		fmt.Println("GET", r.StatusCode, r.Request.URL)
	})

	collector.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		var sb strings.Builder
		sb.WriteString("Redirect: ")

		for _, v := range via {
			sb.WriteString(v.URL.String())
			sb.WriteString(" -> ")
		}

		sb.WriteString(req.URL.String())

		logging.Debug().Println(sb.String())

		// store the redirect so we can get back to the original url later
		scraper.redirects[req.URL.String()] = via[0].URL.String()

		if len(via) > 10 {
			return fmt.Errorf("too many redirects")
		}

		return nil
	})

	collector.OnResponse(func(r *colly.Response) {
		var url string
		var redirected bool

		if scraper.redirects[r.Request.URL.String()] != "" {
			url = scraper.redirects[r.Request.URL.String()]
			redirected = true
		} else {
			url = r.Request.URL.String()
		}

		scraper.scrapedPages[url] = WebPage{
			Url:        url,
			Body:       r.Body,
			Redirected: redirected,
		}
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Fprintf(os.Stderr, "error: %v %s\n", r.StatusCode, err)
	})

	return scraper
}

func (s *Scraper) ScrapeUrls(urls ...string) (map[string]WebPage, error) {
	for _, targetUrl := range urls {
		_, err := url.Parse(targetUrl)

		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse: %s\n", err)
			return nil, err
		}

		s.collector.Visit(targetUrl)
	}

	// make sure async requests have finished
	s.collector.Wait()

	result := map[string]WebPage{}

	for url, webPage := range s.scrapedPages {
		result[url] = webPage
		delete(s.scrapedPages, url)
	}

	return result, nil
}
