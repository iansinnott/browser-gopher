package fulltext_test

import (
	"testing"

	"github.com/iansinnott/browser-gopher/pkg/fulltext"
	"github.com/stretchr/testify/require"
)

// These tests are decidedly not good, in that they depend on the outside world.
// They could fail randomly due to network conditions, DNS issuse, updates in
// the resolved web apps, etc.
func TestScrapeUrls(t *testing.T) {
	table := []struct {
		name string
		url  string
	}{
		{
			name: "scrape a single website",
			url:  "https://example.com",
		},
		{
			name: "handle redirects",
			url:  "https://iansinnott.com", // redirects to https://www.iansinnott.com
		},
	}

	scraper := fulltext.NewScraper()

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			xm, err := scraper.ScrapeUrls(tt.url)
			require.Nil(t, err)
			require.NotEmpty(t, xm[tt.url])
			body := xm[tt.url].Body
			require.NotEmpty(t, body)
		})
	}

	// t.Run("scrape multiple websites", func(t *testing.T) {
	// 	xm, err := scraper.ScrapeUrls("https://example.com", "https://iansinnott.com")
	// 	require.Nil(t, err)
	// 	require.NotEmpty(t, xm["https://example.com"].Body)
	// 	require.NotEmpty(t, xm["https://iansinnott.com"].Body)
	// })
}
