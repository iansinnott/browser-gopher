package fulltext

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/iansinnott/browser-gopher/pkg/logging"
)

type WebPage struct {
	Url        string
	Body       []byte
	Redirected bool
	StatusCode int
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
		colly.MaxDepth(1), // 0 means unlimited. not sure how this actually works since I thought it does NOT spider by default
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
		colly.AllowURLRevisit(),
		// colly.CacheDir(cacheDir), // without cachedir colly will re-request every site (which may be what you want, just note)

		colly.Debugger(&debug.LogDebugger{}),
	)

	err := collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: runtime.NumCPU(),
		Delay:       1,
		RandomDelay: 1,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not set limit rule: %s\n", err)
	}

	scraper := &Scraper{
		collector:    collector,
		scrapedPages: map[string]WebPage{},
		redirects:    map[string]string{},
	}

	collector.OnRequest(func(r *colly.Request) {
		logging.Debug().Println("GET", r.URL)
	})

	collector.OnResponseHeaders(func(r *colly.Response) {
		logging.Debug().Println(r.StatusCode, r.Request.URL)
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
		u, redirected := scraper.UnredirectUrl(r.Request.URL.String())

		scraper.scrapedPages[u] = WebPage{
			Url:        u,
			Body:       r.Body,
			Redirected: redirected,
			StatusCode: r.StatusCode,
		}
	})

	collector.OnError(func(r *colly.Response, err error) {
		logging.Debug().Printf("error: %v %s\n", r.StatusCode, err)
		u, redirected := scraper.UnredirectUrl(r.Request.URL.String())

		scraper.scrapedPages[u] = WebPage{
			Url:        u,
			Body:       r.Body,
			Redirected: redirected,
			StatusCode: r.StatusCode,
		}
	})

	return scraper
}

func (s *Scraper) UnredirectUrl(url string) (u string, redirected bool) {
	if s.redirects[url] != "" {
		return s.redirects[url], true
	}

	return url, false
}

func (s *Scraper) ScrapeUrls(urls ...string) (map[string]WebPage, error) {
	for _, targetUrl := range urls {
		_, err := url.Parse(targetUrl)

		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse: %s\n", err)
			return nil, err
		}

		err = s.collector.Visit(targetUrl)

		if err != nil {
			logging.Debug().Println("could not visit", targetUrl, err)
			return nil, err
		}
	}

	// make sure async requests have finished
	s.collector.Wait()

	// result := map[string]WebPage{}

	// for url, webPage := range s.scrapedPages {
	// 	result[url] = webPage
	// 	delete(s.scrapedPages, url)
	// }

	return s.scrapedPages, nil
}
