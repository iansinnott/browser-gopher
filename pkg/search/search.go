package search

import (
	"context"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/pkg/errors"
)

type Provider interface {
	SearchUrls(query string) ([]types.UrlRow, error)
}

type SearchProvider struct {
	ctx  context.Context
	conf *config.AppConfig
}

func NewSearchProvider(ctx context.Context, conf *config.AppConfig) *SearchProvider {
	return &SearchProvider{ctx: ctx, conf: conf}
}

type URLSearchResult struct {
	Urls  []types.UrlRow
	Count int
}

func (p *SearchProvider) SearchUrls(query string) (*URLSearchResult, error) {
	conn, err := persistence.OpenConnection(p.ctx, p.conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query = "%" + query + "%"

	var count int
	row := conn.QueryRowContext(p.ctx, `
SELECT
	COUNT(*)
FROM
  urls
WHERE
  url LIKE ?
  OR title LIKE ?
  OR description LIKE ?;
	`, query, query, query)
	if row.Err() != nil {
		return nil, errors.Wrap(row.Err(), "row count error")
	}
	err = row.Scan(&count)
	if err != nil {
		return nil, errors.Wrap(err, "row count error")
	}

	rows, err := conn.QueryContext(p.ctx, `
SELECT
  url,
  title,
  description,
  last_visit
FROM
  urls
WHERE
  url LIKE ?
  OR title LIKE ?
  OR description LIKE ?
ORDER BY
  last_visit DESC
LIMIT 100;
	`, query, query, query)

	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "query error")
	}

	xs := []types.UrlRow{}

	for rows.Next() {
		var x types.UrlRow
		var ts int64
		err := rows.Scan(&x.Url, &x.Title, &x.Description, &ts)
		if err != nil {
			return nil, errors.Wrap(err, "row error")
		}
		t := time.Unix(ts, 0)
		x.LastVisit = &t
		xs = append(xs, x)
	}

	return &URLSearchResult{Urls: xs, Count: count}, nil
}
