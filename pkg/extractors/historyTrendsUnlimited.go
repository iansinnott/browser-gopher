package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

// @note We cannot use lastVisitDate in the where clause due to MAX(...) aggregation.
const historyTrendsUrls = `
SELECT
  u.url,
  u.title,
  datetime(max(v.visit_time) / 1e3, 'unixepoch') AS lastVisitDate
FROM
  visits v
  INNER JOIN urls u ON u.urlid = v.urlid
WHERE datetime(v.visit_time / 1e3, 'unixepoch') > ?
GROUP BY
  v.urlid
ORDER BY
  lastVisitDate DESC;
`

const historyTrendsVisits = `
SELECT
  datetime(v.visit_time / 1e3, 'unixepoch') AS visitDate,
  u.url
FROM
  visits v
  INNER JOIN urls u ON u.urlid = v.urlid
WHERE visitDate > ?
ORDER BY 
	visitDate DESC;
`

type HistoryTrendsExtractor struct {
	Name          string
	HistoryDBPath string
}

func (a *HistoryTrendsExtractor) GetName() string {
	return a.Name
}

func (a *HistoryTrendsExtractor) GetDBPath() string {
	return a.HistoryDBPath
}

func (a *HistoryTrendsExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *HistoryTrendsExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM urls;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *HistoryTrendsExtractor) GetAllUrlsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, historyTrendsUrls, since.UTC().Format(util.SQLiteDateTime))
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

func (a *HistoryTrendsExtractor) GetAllVisitsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]types.VisitRow, error) {
	rows, err := conn.QueryContext(ctx, historyTrendsVisits, since.UTC().Format(util.SQLiteDateTime))
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

func FindHistoryTrendsDBs(root string) ([]string, error) {
	results := []string{}

	fmt.Println("Trying root", root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && d.Name() == "1" && strings.Contains(path, "chrome-extension_pnmchffiealhkdloeffcdnbgdnedheme_0") {
			results = append(results, path)
		}
		return nil
	})

	return results, err
}
