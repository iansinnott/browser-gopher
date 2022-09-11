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

const chromiumUrls = `
SELECT
	url,
	title,
	datetime(last_visit_time / 1e6 + strftime('%s', '1601-01-01'), 'unixepoch') as last_visit_date
FROM
	urls;
`

const chromiumVisits = `
SELECT
  datetime(visit_time / 1e6 + strftime ('%s', '1601-01-01'), 'unixepoch') AS visit_date,
  u.url
FROM
  visits v
  INNER JOIN urls u ON v.url = u.id
ORDER BY 
	visit_date DESC;
`

type ChromiumExtractor struct {
	Name          string
	HistoryDBPath string
}

func (a *ChromiumExtractor) GetName() string {
	return a.Name
}

func (a *ChromiumExtractor) GetDBPath() string {
	return a.HistoryDBPath
}

func (a *ChromiumExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *ChromiumExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM urls;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *ChromiumExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, chromiumUrls)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var urls []types.UrlRow

	for rows.Next() {
		var x types.UrlRow
		var visit_time string
		err = rows.Scan(&x.Url, &x.Title, &visit_time)
		if err != nil {
			fmt.Println("individual row error", err)
			return nil, err
		}
		t, err := util.ParseSQLiteDatetime(visit_time)
		if err != nil {
			fmt.Println("could not parse datetime", err)
		}
		x.LastVisit = &t
		urls = append(urls, x)
	}

	err = rows.Err()
	if err != nil {
		fmt.Println("row error", err)
		return nil, err
	}

	return urls, nil
}

func (a *ChromiumExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	rows, err := conn.QueryContext(ctx, chromiumVisits)
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

func FindChromiumDBs(root string) ([]string, error) {
	results := []string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && d.Name() == "History" {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}
