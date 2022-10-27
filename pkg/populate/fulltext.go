package populate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/iansinnott/browser-gopher/pkg/fulltext"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	stripmd "github.com/writeas/go-strip-markdown"
)

// @note the `order by random()` is meant to avoid trying to scrape from the same website all at once. No DoS!
const queryUrlsWithoutDocuments = `
SELECT
	u.url_md5,
  u.url,
	u.title,
	u.description,
  u.last_visit
FROM
  urls u
  LEFT OUTER JOIN documents d ON u.url_md5 = d.url_md5
WHERE
  d.url_md5 IS NULL
ORDER BY 
	RANDOM()
LIMIT ?;
`

func PopulateFulltext(ctx context.Context, db *sql.DB) error {
	limit := 100

	// get all urls that do not have an associated record in the documents table
	rows, err := db.QueryContext(ctx, queryUrlsWithoutDocuments, limit)
	if err != nil {
		return errors.Wrap(err, "failed to query for urls without documents")
	}
	defer rows.Close()

	var urls []types.UrlDbEntity

	// for each document, get the text and index it
	for rows.Next() {
		var u types.UrlDbEntity
		var ts int64

		err := rows.Scan(&u.UrlMd5, &u.Url, &u.Title, &u.Description, &ts)

		// @note last visit time can be zero, indicating unknown visit time. This
		// will happen if importing from browserparrot/persistory because the visits
		// table had a bug
		if ts > 0 {
			t := time.Unix(ts, 0)
			u.LastVisit = &t
		}

		if err != nil {
			return errors.Wrap(err, "error scanning row")
		}

		urls = append(urls, u)
	}

	scraper := fulltext.NewScraper()
	xs := lo.Map(urls, func(u types.UrlDbEntity, i int) string { return u.Url })
	xm, err := scraper.ScrapeUrls(xs...)
	if err != nil {
		return errors.Wrap(err, "error scraping urls")
	}

	for _, u := range urls {
		doc := xm[u.Url]
		converter := md.NewConverter(doc.Url, true, nil)
		md, err := converter.ConvertString(string(doc.Body))
		if err != nil {
			return err
		}

		urlMd5 := util.HashMd5String(doc.Url)
		docMd5 := util.HashMd5String(md) // @note that we use the distilled md hash in order to avoid duplication when content hasn't noticably changed
		markdownPath := filepath.Join("tmp", "markdown", urlMd5+".md")

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`---
url: %s
---

`, doc.Url))
		sb.WriteString(md)

		err = os.WriteFile(markdownPath, []byte(sb.String()), 0644)
		if err != nil {
			return err
		}

		accessedAt := time.Now()
		err = persistence.InsertDocument(ctx, db, &types.DocumentRow{
			DocumentMd5:  docMd5,
			UrlMd5:       urlMd5,
			MarkdownPath: markdownPath,
			StatusCode:   doc.StatusCode,
			AccessedAt:   &accessedAt,
		})
		if err != nil {
			return errors.Wrap(err, "error inserting document")
		}

		// @todo index this plaintext
		plaintext := stripmd.Strip(md)
		searchDoc := SearchableEntity{
			Id:          urlMd5,
			Url:         doc.Url,
			Title:       u.Title,
			Description: u.Description,
			LastVisit:   u.LastVisit,
			Body:        &plaintext,
		}
		err = IndexDocument(ctx, db, searchDoc)
		if err != nil {
			return errors.Wrap(err, "error indexing document")
		}
	}

	return nil
}
