package populate

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
	"github.com/pkg/errors"
	stripmd "github.com/writeas/go-strip-markdown"
)

// how many urls to index at a time
const batchSize = 1000

func BuildIndex(ctx context.Context, db *sql.DB, limit int) (int, error) {
	indexedCount := 0
	toIndexCount, err := persistence.CountUrlsWhere(ctx, db, "indexed_at IS NULL")
	if err != nil {
		return 0, errors.Wrap(err, "error getting count of urls to index")
	}

	if limit > 0 && limit < toIndexCount {
		toIndexCount = limit
	}

	for indexedCount < toIndexCount {
		// get documents to index
		ents, err := getUnindexed(ctx, db)
		if err != nil {
			return 0, err
		}

		// index them
		n, err := batchIndex(ctx, db, ents...)
		if err != nil {
			return 0, err
		}

		// Break out if indexedCount is not increasing
		if n == 0 {
			break
		}

		fmt.Printf("indexing: (%d/%d) %.2f\n", indexedCount, toIndexCount, float32(indexedCount)/float32(toIndexCount))

		indexedCount += n
	}

	return indexedCount, err
}

// Index (or reindex) an individual document. If doc.Id is already present in
// the search index it will be overwritten.
func IndexDocument(ctx context.Context, db *sql.DB, doc types.UrlDbEntity) error {
	_, err := batchIndex(ctx, db, doc)
	if err != nil {
		return errors.Wrap(err, "error indexing document")
	}

	t := time.Now()
	meta := types.UrlMetaRow{
		Url:       doc.Url,
		IndexedAt: &t,
	}

	err = persistence.InsertUrlMeta(ctx, db, meta)
	if err != nil {
		return err
	}

	return nil
}

// The reason we need an int ID is due to the int requirement on rowid in the fts table.
func generateEavId(e string, t string, a string, v string) (int64, error) {
	shasum := util.HashSha1String(fmt.Sprintf("%s%s%s%s", e, t, a, v))

	// take the first 15 characters of the sha1 hash, as this is the max that will
	// fit into sqlite int column (it's signed, i think. otherwise we could take
	// 16 chars)
	hexId := shasum[0:15]

	// convert to a base 10 number
	return strconv.ParseInt(hexId, 16, 64)
}

/**
 * Index an entity
 */
func indexEav(ctx context.Context, db *sql.Tx, e string, t string, a string, v string) error {
	// insert into the fragments table
	const qry = `
		INSERT OR REPLACE INTO
			fragment(id, e, t, a, v)
				VALUES(?, ?, ?, ?, ?);
	`

	id, err := generateEavId(e, t, a, v)
	if err != nil {
		return errors.Wrap(err, "error generating eav id")
	}

	_, err = db.ExecContext(ctx, qry, id, e, t, a, v)
	if err != nil {
		return errors.Wrap(err, "error inserting fragment")
	}

	return nil
}

func batchIndex(ctx context.Context, db *sql.DB, ents ...types.UrlDbEntity) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	cleanupWithError := func(err error) error {
		tx.Rollback()
		return err
	}

	// @todo Could be much more efficient by not committing transaction for each row
	for _, ent := range ents {

		// vs is an array of structs containing `a` and `v` fields
		vs := []struct {
			a string
			v *string
		}{{"url", &ent.Url}, {"title", ent.Title}, {"description", ent.Description}}

		// Insert basic URL data
		for _, x := range vs {
			table := "urls"
			if x.v != nil {
				err := indexEav(ctx, tx, ent.UrlMd5, table, x.a, *x.v)
				if err != nil {
					return 0, cleanupWithError(err)
				}
			}
		}

		// Insert fulltext data
		if ent.Body != nil {
			table := "documents"
			chunk := ""
			// Chunk documents by paragraphs, for now
			for _, paragraph := range strings.Split(*ent.Body, "\n\n") {
				chunk += strings.TrimSpace(paragraph)
				if len(chunk) < 180 {
					chunk += " "
					continue
				}

				err := indexEav(ctx, tx, ent.UrlMd5, table, "content", chunk)
				if err != nil {
					return 0, cleanupWithError(err)
				}

				chunk = ""
			}

			// Grab straggling chunk. i.e. if the whole document is less than the threshold.
			if len(chunk) > 0 {
				err := indexEav(ctx, tx, ent.UrlMd5, table, "content", chunk)
				if err != nil {
					return 0, cleanupWithError(err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	// var err error

	metas := []types.UrlMetaRow{}

	// Mark docs as indexed so that we don't re-index them
	for _, doc := range ents {
		t := time.Now()
		metas = append(metas, types.UrlMetaRow{
			Url:       doc.Url,
			IndexedAt: &t,
		})
	}

	err = persistence.InsertUrlMeta(ctx, db, metas...)

	if err != nil {
		// check if SQLITE_BUSY error
		if strings.Contains(err.Error(), "database is locked") {
			fmt.Println("database is locked, retrying...")
			time.Sleep(1000 * time.Millisecond)
			return batchIndex(ctx, db, ents...)
		}

		return 0, errors.Wrap(err, "error marking doc as indexed")
	}

	return len(ents), nil
}

func getUnindexed(ctx context.Context, db *sql.DB) ([]types.UrlDbEntity, error) {
	const qry = `
		SELECT
			u.url_md5,
			u.url,
			u.title,
			u.description,
			u.last_visit,
			doc.document_md5,
			doc.body
		FROM
			urls u
			LEFT OUTER JOIN urls_meta um ON u.url_md5 = um.url_md5
			LEFT OUTER JOIN url_document_edges ON u.url_md5 = url_document_edges.url_md5
			LEFT OUTER JOIN documents doc ON url_document_edges.document_md5 = doc.document_md5
		WHERE
			um.indexed_at IS NULL
		ORDER BY 
			last_visit DESC
		LIMIT ?;
	`

	rows, err := db.QueryContext(ctx, qry, batchSize)
	if err != nil {
		return nil, errors.Wrap(err, "error querying unindexed urls")
	}
	defer rows.Close()

	// Put docs into a slice so that we can iterate over them to mark them as
	// indexed. Otherwies we could add them to the batch directly.
	var docs []types.UrlDbEntity

	for rows.Next() {
		var ent types.UrlDbEntity
		var ts int64
		err := rows.Scan(&ent.UrlMd5, &ent.Url, &ent.Title, &ent.Description, &ts, &ent.BodyMd5, &ent.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}

		// @note last visit time can be zero, indicating unknown visit time. This
		// will happen if importing from browserparrot/persistory because the visits
		// table had a bug
		if ts > 0 {
			t := time.Unix(ts, 0)
			ent.LastVisit = &t
		}

		if ent.Body != nil {
			plaintext := stripmd.Strip(*ent.Body)
			ent.Body = &plaintext
		}

		docs = append(docs, ent)
	}

	return docs, nil
}

func ReindexWithLimit(ctx context.Context, db *sql.DB, limit int) (int, error) {
	var err error
	qry := `
		UPDATE
			urls_meta
		SET
			indexed_at = NULL
		WHERE
			indexed_at NOT NULL;
	`
	_, err = db.ExecContext(ctx, qry)
	if err != nil {
		return 0, errors.Wrap(err, "error removing indexed status")
	}

	return BuildIndex(ctx, db, limit)
}

// Reindex documents that have already been indexed. This does not remove
// anything from the index, but will overwrite documents that have been updated.
func ReindexAll(ctx context.Context, db *sql.DB) (int, error) {
	return ReindexWithLimit(ctx, db, 0)
}
