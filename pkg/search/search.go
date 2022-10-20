package search

import (
	"github.com/iansinnott/browser-gopher/pkg/types"
)

type URLQueryResult struct {
	Urls  []types.UrlDbEntity
	Count uint
}

type SearchProvider interface {
	SearchUrls(query string) (*URLQueryResult, error)
}

type DataProvider interface {
	SearchProvider
	RecentUrls(limit uint) (*URLQueryResult, error)
}
