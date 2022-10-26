package fulltext

import (
	"fmt"
	"net/url"
	"os"

	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	collector   *colly.Collector
	scrapedUrls map[string][]byte
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
		colly.Async(false),
		colly.IgnoreRobotsTxt(),
	)

	scraper := &Scraper{
		collector:   collector,
		scrapedUrls: map[string][]byte{},
	}

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Fetching", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		l := r.Headers.Get("Content-Length")
		if l == "" {
			fmt.Println("No content length header")
		} else {
			fmt.Println("Fetched", l, "bytes")
		}
		scraper.scrapedUrls[r.Request.URL.String()] = r.Body
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Fprintf(os.Stderr, "error: %v %s\n", r.StatusCode, err)
	})

	return scraper
}

func (s *Scraper) ScrapeUrls(urls []string) (map[string][]byte, error) {
	for _, targetUrl := range urls {
		_, err := url.Parse(targetUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse: %s\n", err)
			return nil, err
		}

		// hostname = u.Hostname()
		// pathname = strings.ReplaceAll(strings.Trim(u.Path, "/"), "/", "_")

		s.collector.Visit(targetUrl)
	}

	// make sure async requests have finished
	s.collector.Wait()

	result := map[string][]byte{}

	for url, body := range s.scrapedUrls {
		result[url] = body
		delete(s.scrapedUrls, url)
	}

	return result, nil
}
