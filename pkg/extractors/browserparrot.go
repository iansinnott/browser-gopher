package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/iansinnott/browser-gopher/pkg/types"
)

const browserParrotUrls = `
SELECT
  url,
  title
FROM
  datasource_browsing_history;
`

type BrowserParrotExtractor struct {
	Name          string
	HistoryDBPath string
}

func (a *BrowserParrotExtractor) GetName() string {
	return a.Name
}

func (a *BrowserParrotExtractor) GetDBPath() string {
	return a.HistoryDBPath
}
func (a *BrowserParrotExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *BrowserParrotExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM datasource_browsing_history;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *BrowserParrotExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, browserParrotUrls)
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

// Persistory / browser parrot does not map visits properly as of this commit
func (a *BrowserParrotExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	return []types.VisitRow{}, nil
}
