package populate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	_ "github.com/blevesearch/bleve/v2/analysis/lang/cjk"
	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/pkg/errors"
	stripmd "github.com/writeas/go-strip-markdown"
)

// reference the index so that it does't get opened multiple times (causes
// breakage). What's the best way to do this?
var index bleve.Index

func GetIndex() (*bleve.Index, error) {
	if index != nil {
		return &index, nil
	}

	var (
		idx bleve.Index
		err error
	)

	// if bleve index exists on disk then open it, otherwise create a new index
	_, err = os.Stat(config.Config.SearchIndexPath)

	if os.IsNotExist(err) {
		// start building up the doc mapping
		docMapping := bleve.NewDocumentMapping()

		// for md5s, which are used as id fields, do not index them. but we need them stored for retreival later
		md5Mapping := bleve.NewTextFieldMapping()
		md5Mapping.Store = true
		md5Mapping.Index = false
		docMapping.AddFieldMappingsAt("url_md5", md5Mapping)
		docMapping.AddFieldMappingsAt("body_md5", md5Mapping)

		// handle datetimes
		lastVisitMapping := bleve.NewDateTimeFieldMapping()
		docMapping.AddFieldMappingsAt("last_visit", lastVisitMapping)

		urlMapping := bleve.NewTextFieldMapping()
		urlMapping.Store = true
		urlMapping.Index = true
		// There is also 'web', supposedly: https://bleveanalysis.couchbase.com/analysis
		urlMapping.Analyzer = "simple" // docs: https://blevesearch.com/docs/Analyzers/
		docMapping.AddFieldMappingsAt("url", urlMapping)

		// Text mapping uses CJK because it also supports latin script. The 'en' analyzer (or other language)
		// might be more optimal for a specific language, but in the interest of supporting non-latin scripts
		// we're using cjk. See: https://bleveanalysis.couchbase.com/analysis for how these analyzers differ.
		// Have not tested with other non-spaced, non-latin languages (such as Thai).
		textMapping := bleve.NewTextFieldMapping()
		textMapping.Store = true
		textMapping.Index = true
		textMapping.Analyzer = "cjk"
		docMapping.AddFieldMappingsAt("title", textMapping)
		docMapping.AddFieldMappingsAt("description", textMapping)

		// the document body will be stored in sqlite, so does not need to be stored here
		bodyMapping := bleve.NewTextFieldMapping()
		bodyMapping.Store = true // @todo We can have a smaller index by not storing this, but then we need to hit the db for found items to get the full body text.
		bodyMapping.Index = true
		textMapping.Analyzer = "cjk" // see textMapping for details
		docMapping.AddFieldMappingsAt("body", bodyMapping)

		mapping := bleve.NewIndexMapping()
		mapping.DefaultMapping = docMapping
		idx, err = bleve.New(config.Config.SearchIndexPath, mapping)
		if err != nil {
			err = errors.Wrap(err, "error creating new index")
		}
	} else {
		idx, err = bleve.Open(config.Config.SearchIndexPath)
		if err != nil {
			err = errors.Wrap(err, "error opening index")
		}
	}

	index = idx

	return &idx, err
}

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

	idx, err := GetIndex()
	if err != nil {
		return 0, errors.Wrap(err, "error getting index")
	}

	for indexedCount < toIndexCount {
		// get documents to index
		ents, err := getUnindexed(ctx, db)
		if err != nil {
			return 0, err
		}

		// index them
		n, err := batchIndex(ctx, db, *idx, ents...)
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
func IndexDocument(ctx context.Context, db *sql.DB, doc types.SearchableEntity) error {
	idx, err := GetIndex()
	if err != nil {
		return errors.Wrap(err, "error getting index")
	}

	err = (*idx).Index(doc.Id, doc)
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

func batchIndex(ctx context.Context, db *sql.DB, idx bleve.Index, ents ...types.SearchableEntity) (int, error) {
	batch := idx.NewBatch()
	for _, ent := range ents {
		batch.Index(ent.Id, ent)
	}

	err := idx.Batch(batch)
	if err != nil {
		return 0, errors.Wrap(err, "batch error")
	}

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
			return batchIndex(ctx, db, idx, ents...)
		}

		return 0, errors.Wrap(err, "error marking doc as indexed")
	}

	return len(ents), nil
}

func getUnindexed(ctx context.Context, db *sql.DB) ([]types.SearchableEntity, error) {
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
	var docs []types.SearchableEntity

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

		doc := types.SearchableEntity{
			Id:          ent.UrlMd5,
			Url:         ent.Url,
			Title:       ent.Title,
			Description: ent.Description,
			LastVisit:   ent.LastVisit,
			Body:        ent.Body,
			BodyMd5:     ent.BodyMd5,
		}

		if doc.Body != nil {
			plaintext := stripmd.Strip(*ent.Body)
			doc.Body = &plaintext
		}

		docs = append(docs, doc)
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
