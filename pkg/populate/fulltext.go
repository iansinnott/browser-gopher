package populate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/iansinnott/browser-gopher/pkg/fulltext"
	"github.com/iansinnott/browser-gopher/pkg/logging"
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
  LEFT OUTER JOIN url_document_edges edge ON u.url_md5 = edge.url_md5
WHERE
  edge.url_md5 IS NULL
ORDER BY 
	RANDOM()
LIMIT ?;
`

const countUrlsWithoutDocuments = `
SELECT
	COUNT(*)
FROM
  urls u
  LEFT OUTER JOIN url_document_edges edge ON u.url_md5 = edge.url_md5
WHERE
  edge.url_md5 IS NULL;
`

const scrapeBatchSize = 100

func PopulateFulltext(ctx context.Context, db *sql.DB) (int, error) {
	indexedCount := 0
	var todoCount int
	row := db.QueryRowContext(ctx, countUrlsWithoutDocuments)
	err := row.Scan(&todoCount)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count urls without documents")
	}

	scraper := fulltext.NewScraper()

	// Do the scraping
	for indexedCount < todoCount {
		fmt.Printf("scraping: (%d/%d) %.2f\n", indexedCount, todoCount, float32(indexedCount)/float32(todoCount))

		n, err := batchScrape(ctx, db, scraper)

		// break early if there was an error
		if err != nil {
			return 0, err
		}

		// Break early if we get back fewer URLs than batch size, indicating there
		// are less than batch size left to scrape
		if n == 0 {
			break
		}

		indexedCount += n
	}

	idx, err := GetIndex()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get index")
	}
	indexedCount = 0
	toIndexCount, err := persistence.CountUrlsWhere(ctx, db,
		`documents.body NOT NULL 
			AND documents.body != '' 
			AND urls_meta.indexed_at < documents.accessed_at`)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count urls to index")
	}

	// Do the indexing
	for indexedCount < toIndexCount {
		ents, err := getUnindexedBodyRows(ctx, db)
		if err != nil {
			return 0, errors.Wrap(err, "error getting unindexed bodies")
		}

		n, err := batchIndex(ctx, db, *idx, ents...)
		if err != nil {
			return 0, errors.Wrap(err, "error indexing batch")
		}

		if n == 0 {
			logging.Debug().Println("nothing left to process")
			break
		}

		logging.Debug().Printf("indexing with bodies (%d/%d) %.2f\n", indexedCount, toIndexCount, float32(indexedCount)/float32(toIndexCount))
		indexedCount += n
	}

	return indexedCount, nil
}

func getUnindexedBodyRows(ctx context.Context, db *sql.DB) ([]types.SearchableEntity, error) {
	qry := `
			SELECT
				u.url_md5,
				u.url,
				u.title,
				u.description,
				u.last_visit,
				d.body
			FROM
				urls u
				JOIN url_document_edges edge ON u.url_md5 = edge.url_md5
				JOIN documents d ON edge.document_md5 = d.document_md5
				JOIN urls_meta m ON u.url_md5 = m.url_md5
			WHERE
				d.body NOT NULL
				AND d.body != ''
				AND m.indexed_at < d.accessed_at
			LIMIT ?;
		`

	rows, err := db.QueryContext(ctx, qry, batchSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for urls with unindexed documents")
	}
	defer rows.Close()

	var ents []types.SearchableEntity

	for rows.Next() {
		var (
			ent types.SearchableEntity
			ts  int64
			t   time.Time
		)

		err := rows.Scan(
			&ent.Id,
			&ent.Url,
			&ent.Title,
			&ent.Description,
			&ts,
			&ent.Body,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		if ts == 0 {
			t = time.Unix(ts, 0)
			ent.LastVisit = &t
		}

		// @todo index this plaintext. Will do this from our data base, since fulltext is now stored there
		plaintext := stripmd.Strip(*ent.Body)
		ent.Body = &plaintext

		ents = append(ents, ent)
	}

	return ents, nil
}

func batchScrape(ctx context.Context, db *sql.DB, scraper *fulltext.Scraper) (int, error) {
	// get all urls that do not have an associated record in the documents table
	rows, err := db.QueryContext(ctx, queryUrlsWithoutDocuments, scrapeBatchSize)
	if err != nil {
		return 0, errors.Wrap(err, "failed to query for urls without documents")
	}

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
			return 0, errors.Wrap(err, "error scanning row")
		}

		urls = append(urls, u)
	}

	rows.Close()

	xs := lo.Map(urls, func(u types.UrlDbEntity, i int) string { return u.Url })
	xm, err := scraper.ScrapeUrls(xs...)
	if err != nil {
		return 0, errors.Wrap(err, "error scraping urls")
	}

	for _, u := range urls {
		doc := xm[u.Url]
		converter := md.NewConverter(doc.Url, true, nil)
		md, err := converter.ConvertString(string(doc.Body))
		if err != nil {
			return 0, err
		}

		urlMd5 := util.HashMd5String(doc.Url)
		docMd5 := util.HashMd5String(md) // @note that we use the distilled md hash in order to avoid duplication when content hasn't noticably changed
		accessedAt := time.Now()

		err = persistence.InsertDocument(ctx, db, &types.DocumentRow{
			DocumentMd5: docMd5,
			UrlMd5:      urlMd5,
			StatusCode:  doc.StatusCode,
			AccessedAt:  &accessedAt,
			Body:        &md,
		})
		if err != nil {
			return 0, errors.Wrap(err, "error inserting document")
		}

	}

	return len(urls), nil
}
