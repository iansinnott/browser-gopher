package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

const orionUrls = `
SELECT
  url,
  title
FROM
  history_items;
`

const orionVisits = `
SELECT
  v.VISIT_TIME,
  u.URL
FROM
  visits v
  INNER JOIN history_items u ON u.ID = v.HISTORY_ITEM_ID
ORDER BY
  VISIT_TIME DESC;
`

type OrionExtractor struct {
	Name          string
	HistoryDBPath string
}

func (a *OrionExtractor) GetName() string {
	return a.Name
}

func (a *OrionExtractor) GetDBPath() string {
	return a.HistoryDBPath
}

func (a *OrionExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *OrionExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM history_items;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *OrionExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, orionUrls)
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

func (a *OrionExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	rows, err := conn.QueryContext(ctx, orionVisits)
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

		t, err := util.ParseISODatetime(ts)
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
