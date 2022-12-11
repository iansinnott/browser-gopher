package fulltext

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
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
	lock         sync.RWMutex
}

// currently unused
func (s *Scraper) onRequest(r *colly.Request) {}

func (s *Scraper) redirectHandler(req *http.Request, via []*http.Request) error {
	var sb strings.Builder
	sb.WriteString("Redirect: ")
	for _, v := range via {
		sb.WriteString(v.URL.String())
		sb.WriteString(" -> ")
	}
	sb.WriteString(req.URL.String())
	logging.Debug().Println(sb.String())

	// store the redirect so we can get back to the original url later
	s.lock.Lock()
	s.redirects[req.URL.String()] = via[0].URL.String()
	s.lock.Unlock()

	if len(via) > 10 {
		return fmt.Errorf("too many redirects")
	}

	return nil
}

func (s *Scraper) handleResponse(r *colly.Response) {
	u, redirected := s.UnredirectUrl(r.Request.URL.String())

	s.lock.Lock()
	defer s.lock.Unlock()

	s.scrapedPages[u] = WebPage{
		Url:        u,
		Body:       r.Body,
		Redirected: redirected,
		StatusCode: r.StatusCode,
	}
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
	)

	// if logging.IsDebug() {
	// 	collector.SetDebugger(&debug.LogDebugger{})
	// }

	perSiteConcurrency := 2
	logging.Debug().Printf("setting max concurrency to %d", perSiteConcurrency)
	err := collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: perSiteConcurrency,
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

	collector.OnRequest(scraper.onRequest)
	collector.OnResponseHeaders(func(r *colly.Response) {
		logging.Debug().Println(r.StatusCode, r.Request.URL)
	})
	collector.SetRedirectHandler(scraper.redirectHandler)
	collector.OnResponse(scraper.handleResponse)

	collector.OnError(func(r *colly.Response, err error) {
		if r.StatusCode < 400 {
			logging.Debug().Printf("error: %v %s\n", r.StatusCode, err)
		}
		scraper.handleResponse(r)
	})

	return scraper
}

func (s *Scraper) UnredirectUrl(url string) (u string, redirected bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.redirects[url] != "" {
		return s.redirects[url], true
	}

	return url, false
}

func (s *Scraper) ScrapeUrls(urls ...string) (map[string]WebPage, error) {
	for _, targetUrl := range urls {
		_, err := url.Parse(targetUrl)

		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: could not parse: %s\n", err)
			continue
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
