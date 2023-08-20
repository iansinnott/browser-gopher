package search

import (
	"github.com/iansinnott/browser-gopher/pkg/types"
)

type URLQueryResult struct {
	Urls  []types.UrlDbEntity
	Count uint
}

type SearchResult struct {
	Urls  []types.SearchableEntity
	Count uint
}

type SearchProvider interface {
	SearchUrls(query string) (*SearchResult, error)
}

type DataProvider interface {
	SearchProvider
	RecentUrls(limit uint) (*SearchResult, error)
}
