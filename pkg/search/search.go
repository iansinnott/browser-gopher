package search

import (
	"github.com/iansinnott/browser-gopher/pkg/types"
)

type URLQueryResult struct {
	Urls  []types.UrlRow
	Count int
}

type SearchProvider interface {
	SearchUrls(query string) (*URLQueryResult, error)
}

type DataProvider interface {
	SearchProvider
	RecentUrls(limit int) (*URLQueryResult, error)
}
