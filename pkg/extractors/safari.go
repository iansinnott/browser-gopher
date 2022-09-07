package extractors

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/iansinnott/browser-gopher/pkg/types"
)

type SafariExtractor struct {
	Name          string
	HistoryDBPath string
}

// Join with latest visit to get title, since safari doesn't store title with URL
const safariUrls = `
SELECT
  u.url AS url,
  v.title AS title
FROM
  history_items u
  INNER JOIN (
    SELECT
      *,
      max(visit_time) AS last_visit_date
    FROM
      history_visits
    GROUP BY
      history_item) v ON v.history_item = u.id;
`

const safariVisits = `
SELECT
  datetime(visit_time + 978307200, 'unixepoch') AS visit_date,
  u.url
FROM
  history_visits v
  INNER JOIN history_items u ON v.history_item = u.id;
`

func (a *SafariExtractor) GetName() string {
	return a.Name
}

func (a *SafariExtractor) GetDBPath() string {
	return a.HistoryDBPath
}

func (a *SafariExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, safariUrls)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var urls []types.UrlRow

	for rows.Next() {
		var x types.UrlRow
		err = rows.Scan(&x.Url, &x.Title)
		if err != nil {
			fmt.Println("individual row error", err)
			return nil, err
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

func (a *SafariExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	return []types.VisitRow{}, nil
}
