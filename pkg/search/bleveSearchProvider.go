package search

import (
	"context"

	"github.com/iansinnott/browser-gopher/pkg/config"
)

type BleveSearchProvider struct {
	ctx  context.Context
	conf *config.AppConfig
}

func NewBleveSearchProvider(ctx context.Context, conf *config.AppConfig) SqlSearchProvider {
	return SqlSearchProvider{ctx: ctx, conf: conf}
}

func (p BleveSearchProvider) SearchUrls(query string) (*URLQueryResult, error) {
	// @todo Implement
	return nil, nil
}
