package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/iansinnott/browser-gopher/pkg/types"
)

type URLQueryResult struct {
	Urls  []types.UrlDbEntity
	Count uint
	Meta  *bleve.SearchResult
}

type SearchResult struct {
	Urls  []types.SearchableEntity
	Count uint
	Meta  *bleve.SearchResult
}

type SearchProvider interface {
	SearchUrls(query string) (*SearchResult, error)
}

type DataProvider interface {
	SearchProvider
	RecentUrls(limit uint) (*SearchResult, error)
}
