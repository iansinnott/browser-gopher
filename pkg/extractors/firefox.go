package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

type FirefoxExtractor struct {
	Name          string
	HistoryDBPath string
}

const firefoxUrls = `
SELECT
	url,
	title,
	description,
  datetime(last_visit_date / 1e6, 'unixepoch') as last_visit_date
FROM
	moz_places;
`

const firefoxVisits = `
SELECT
  datetime(v.visit_date / 1e6, 'unixepoch') AS visit_date,
  u.url
FROM
  moz_historyvisits v
  INNER JOIN moz_places u ON v.place_id = u.id;
`

func (a *FirefoxExtractor) GetName() string {
	return a.Name
}

func (a *FirefoxExtractor) GetDBPath() string {
	return a.HistoryDBPath
}
func (a *FirefoxExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *FirefoxExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM moz_places;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *FirefoxExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, firefoxUrls)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var urls []types.UrlRow

	for rows.Next() {
		var x types.UrlRow
		var visit_time *string
		err = rows.Scan(&x.Url, &x.Title, &x.Description, &visit_time)
		if err != nil {
			fmt.Println("individual row error", err)
			return nil, err
		}
		if visit_time != nil {
			t, err := util.ParseSQLiteDatetime(*visit_time)
			if err != nil {
				fmt.Println("could not parse datetime", err)
			}
			x.LastVisit = &t
		}

		urls = append(urls, x)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println("row error", err)
		return nil, err
	}

	return urls, nil
}

func (a *FirefoxExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	rows, err := conn.QueryContext(ctx, firefoxVisits)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var visits []types.VisitRow

	for rows.Next() {
		var x types.VisitRow
		var ts string
		err = rows.Scan(&ts, &x.Url)
		if err != nil {
			fmt.Println("individual row error", err)
			return nil, err
		}

		t, err := util.ParseSQLiteDatetime(ts)
		if err != nil {
			fmt.Println("datetime parsing error", ts, err)
			return nil, err
		}
		x.Datetime = t
		visits = append(visits, x)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println("row error", err)
		return nil, err
	}

	return visits, nil
}

func FindFirefoxDBs(root string) ([]string, error) {
	results := []string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.Name() == "places.sqlite" {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}
