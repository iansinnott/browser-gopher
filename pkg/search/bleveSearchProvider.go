package search

import (
	"context"

	"github.com/blevesearch/bleve/v2"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/populate"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/samber/lo"
)

type BleveSearchProvider struct {
	ctx  context.Context
	conf *config.AppConfig
}

func NewBleveSearchProvider(ctx context.Context, conf *config.AppConfig) BleveSearchProvider {
	return BleveSearchProvider{ctx: ctx, conf: conf}
}

// Search the Bleve index. Fields array can be used to specify which fields for
// which to return the full text value. pass []string{"*"] for all fields. defaults to
// empty. Any matched fields will be at least partially included via the fragments struct
func (p BleveSearchProvider) SearchBleve(query string, fields ...string) (*bleve.SearchResult, error) {
	qry := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequest(qry)
	req.Fields = fields
	req.Size = 100 // item count
	req.From = 0   // for pagination
	req.IncludeLocations = true
	req.Explain = false                  // could be useful in the future
	req.Highlight = bleve.NewHighlight() // highlight results. by default with <mark> tags

	idx, err := populate.GetIndex()
	if err != nil {
		return nil, err
	}

	return (*idx).Search(req)
}

func (p BleveSearchProvider) SearchUrls(query string) (*SearchResult, error) {
	result, err := p.SearchBleve(query)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(result.Hits))
	for i, hit := range result.Hits {
		ids[i] = hit.ID
	}

	conn, err := persistence.OpenConnection(p.ctx, p.conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	xs, err := persistence.UrlsById(p.ctx, conn, ids...)
	if err != nil {
		return nil, err
	}

	searchResult := lo.Map(xs, func(x types.UrlDbEntity, i int) types.SearchableEntity {
		return types.UrlDbEntityToSearchableEntity(x)
	})

	return &SearchResult{Urls: searchResult, Count: uint(result.Total), Meta: result}, err
}
