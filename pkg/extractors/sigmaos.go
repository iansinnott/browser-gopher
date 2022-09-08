package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

const sigmaUrls = `
SELECT
  u.ZURL AS url,
  v.ZTITLE AS title
FROM
  ZHISTORYITEM u
  INNER JOIN ZHISTORYVISIT v ON u.Z_PK = v.ZHISTORYITEM;
`

const sigmaVisits = `
SELECT
  datetime (v.ZVISITTIME + 978307200, 'unixepoch') AS visit_time,
  u.ZURL AS url
FROM
  ZHISTORYITEM u
  INNER JOIN ZHISTORYVISIT v ON u.Z_PK = v.ZHISTORYITEM
ORDER BY
  v.ZVISITTIME DESC;
`

type SigmaOSExtractor struct {
	Name          string
	HistoryDBPath string
}

func (a *SigmaOSExtractor) GetName() string {
	return a.Name
}

func (a *SigmaOSExtractor) GetDBPath() string {
	return a.HistoryDBPath
}

func (a *SigmaOSExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM ZHISTORYITEM;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *SigmaOSExtractor) GetAllUrls(ctx context.Context, conn *sql.DB) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, sigmaUrls)
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

func (a *SigmaOSExtractor) GetAllVisits(ctx context.Context, conn *sql.DB) ([]types.VisitRow, error) {
	rows, err := conn.QueryContext(ctx, sigmaVisits)
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
