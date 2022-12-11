package search

import (
	"context"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type SqlSearchProvider struct {
	ctx  context.Context
	conf *config.AppConfig
}

func NewSqlSearchProvider(ctx context.Context, conf *config.AppConfig) SqlSearchProvider {
	return SqlSearchProvider{ctx: ctx, conf: conf}
}

func (p SqlSearchProvider) SearchUrls(query string) (*SearchResult, error) {
	conn, err := persistence.OpenConnection(p.ctx, p.conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query = "%" + query + "%"

	var count uint
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
	url_md5,
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

	xs := []types.UrlDbEntity{}

	for rows.Next() {
		var x types.UrlDbEntity
		var ts int64
		err := rows.Scan(&x.UrlMd5, &x.Url, &x.Title, &x.Description, &ts)
		if err != nil {
			return nil, errors.Wrap(err, "row error")
		}
		t := time.Unix(ts, 0)
		x.LastVisit = &t
		xs = append(xs, x)
	}

	searchResult := lo.Map(xs, func(x types.UrlDbEntity, i int) types.SearchableEntity {
		return types.UrlDbEntityToSearchableEntity(x)
	})

	return &SearchResult{Urls: searchResult, Count: count}, nil
}

func (p SqlSearchProvider) RecentUrls(limit uint) (*SearchResult, error) {
	conn, err := persistence.OpenConnection(p.ctx, p.conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var count uint
	row := conn.QueryRowContext(p.ctx, `
SELECT
	COUNT(*)
FROM
  urls;
	`)
	if row.Err() != nil {
		return nil, errors.Wrap(row.Err(), "row count error")
	}
	err = row.Scan(&count)
	if err != nil {
		return nil, errors.Wrap(err, "row count error")
	}

	rows, err := conn.QueryContext(p.ctx, `
SELECT
	url_md5,
  url,
  title,
  description,
  last_visit
FROM
  urls
ORDER BY
  last_visit DESC
LIMIT ?;
	`, limit)

	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "query error")
	}

	xs := []types.UrlDbEntity{}

	for rows.Next() {
		var x types.UrlDbEntity
		var ts int64
		err := rows.Scan(&x.UrlMd5, &x.Url, &x.Title, &x.Description, &ts)
		if err != nil {
			return nil, errors.Wrap(err, "row error")
		}
		t := time.Unix(ts, 0)
		x.LastVisit = &t
		xs = append(xs, x)
	}

	searchResult := lo.Map(xs, func(x types.UrlDbEntity, i int) types.SearchableEntity {
		return types.UrlDbEntityToSearchableEntity(x)
	})

	return &SearchResult{Urls: searchResult, Count: count}, nil
}
