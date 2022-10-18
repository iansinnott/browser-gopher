package populate

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/pkg/errors"
)

const queryUnindexedUrls = `
SELECT
  url_md5,
  url,
  title,
  description,
  last_visit
FROM urls
WHERE indexed_at IS NULL
LIMIT 100
`

type SearchableEntity struct {
	Id  string
	Url string

	// Nullable types later
	// Title       string
	// Description string
}

func GetIndex() (bleve.Index, error) {
	var (
		idx bleve.Index
		err error
	)

	// if index.bleve exists on disk then open it, otherwise create a new index
	_, err = os.Stat("index.bleve")

	if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		idx, err = bleve.New("index.bleve", mapping)
		if err != nil {
			err = errors.Wrap(err, "error creating new index")
		}
	} else {
		idx, err = bleve.Open("index.bleve")
		if err != nil {
			err = errors.Wrap(err, "error opening index")
		}
	}

	return idx, err
}

func BuildIndex(ctx context.Context, db *sql.DB) error {
	idx, err := GetIndex()
	if err != nil {
		return errors.Wrap(err, "error getting index")
	}

	batch := idx.NewBatch()

	rows, err := db.QueryContext(ctx, queryUnindexedUrls)
	if err != nil {
		return errors.Wrap(err, "error querying unindexed urls")
	}
	defer rows.Close()

	// Put docs into a slice so that we can iterate over them to mark them as
	// indexed. Otherwies we could add them to the batch directly.
	var docs []SearchableEntity

	for rows.Next() {
		var ent types.UrlDbEntity
		var ts int64
		err := rows.Scan(&ent.UrlMd5, &ent.Url, &ent.Title, &ent.Description, &ts)
		if err != nil {
			return errors.Wrap(err, "error scanning row")
		}

		t := time.Unix(ts, 0)
		ent.LastVisit = &t

		doc := SearchableEntity{
			Id:  ent.UrlMd5,
			Url: ent.Url,
		}

		docs = append(docs, doc)
	}

	for _, doc := range docs {
		batch.Index(doc.Id, doc)
	}

	err = idx.Batch(batch)
	if err != nil {
		return errors.Wrap(err, "batch error")
	}

	// Mark docs as indexed
	for _, doc := range docs {
		t := time.Now()
		meta := &types.UrlMetaRow{
			Url:       doc.Url,
			IndexedAt: &t,
		}

		err := persistence.InsertUrlMeta(ctx, db, meta)

		if err != nil {
			return errors.Wrap(err, "error marking doc as indexed")
		}
	}

	return nil
}
