package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

const sigmaUrls = `
SELECT
  u.ZURL AS url,
  v.ZTITLE AS title,
	DATETIME(MAX(v.ZVISITTIME) + 978307200, 'unixepoch') AS visit_time
FROM
  ZHISTORYITEM u
  INNER JOIN ZHISTORYVISIT v ON u.Z_PK = v.ZHISTORYITEM
GROUP BY v.ZHISTORYITEM
ORDER BY
  v.ZVISITTIME DESC;
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
func (a *SigmaOSExtractor) SetDBPath(s string) {
	a.HistoryDBPath = s
}

func (a *SigmaOSExtractor) VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error) {
	row := conn.QueryRowContext(ctx, "SELECT count(*) FROM ZHISTORYITEM;")
	err := row.Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *SigmaOSExtractor) GetAllUrlsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]types.UrlRow, error) {
	rows, err := conn.QueryContext(ctx, sigmaUrls)
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

func (a *SigmaOSExtractor) GetAllVisitsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]types.VisitRow, error) {
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
