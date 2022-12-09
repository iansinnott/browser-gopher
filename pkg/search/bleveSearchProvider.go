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

func (p BleveSearchProvider) SearchBleve(query string) (*bleve.SearchResult, error) {
	qry := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequest(qry)
	req.Size = 100
	req.Fields = append(req.Fields, "id", "url", "title", "description", "last_visit", "body")
	req.IncludeLocations = true

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
