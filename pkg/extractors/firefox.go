package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/iansinnott/browser-gopher/pkg/types"
)

type FirefoxExtractor struct {
	Name          string
	HistoryDBPath string
}

const firefoxUrls = `
SELECT
	url,
	title,
	description
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

func (a *FirefoxExtractor) GetAllUrls() ([]types.UrlRow, error) {
	log.Println("["+a.Name+"] reading", a.HistoryDBPath)
	conn, err := sql.Open("sqlite", a.HistoryDBPath)

	if err != nil {
		fmt.Println("could not connect to db at", a.HistoryDBPath, err)
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(context.TODO(), firefoxUrls)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var urls []types.UrlRow

	for rows.Next() {
		var x types.UrlRow
		err = rows.Scan(&x.Url, &x.Title, &x.Description)
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

func (a *FirefoxExtractor) GetAllVisits() ([]types.VisitRow, error) {
	return []types.VisitRow{}, nil
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
