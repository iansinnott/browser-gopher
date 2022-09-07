package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	"github.com/iansinnott/browser-gopher/pkg/types"
)

type SafariExtractor struct {
	Name          string
	HistoryDBPath string
}

// Join with latest visit to get title, since safari doesn't store title with URL
const queryUrls = `
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

func (a *SafariExtractor) GetName() string {
	return a.Name
}

func (a *SafariExtractor) GetAllUrls() ([]types.UrlRow, error) {
	log.Println("["+a.Name+"] reading", a.HistoryDBPath)
	conn, err := sql.Open("sqlite", a.HistoryDBPath)

	if err != nil {
		fmt.Println("could not connect to db at", a.HistoryDBPath, err)
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(context.TODO(), queryUrls)
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

func (a *SafariExtractor) GetAllVisits() ([]types.VisitRow, error) {
	return []types.VisitRow{}, nil
}
