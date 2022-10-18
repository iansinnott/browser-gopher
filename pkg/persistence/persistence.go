package persistence

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

// @note Initially visits had a unique index on `extractor_name, url_md5,
// visit_time`, however, this lead to duplicate visits. The visits were
// duplicated because some browsers will immport the history of other browsers,
// or in cases like the history trends chrome extension duplication is
// explicitly part of the goal. Thus, in order to minimize duplication visits
// are considered unique by url and unix timestamp.
const initSql = `
CREATE TABLE IF NOT EXISTS "urls" (
  "url_md5" VARCHAR(32) PRIMARY KEY NOT NULL,
  "url" TEXT UNIQUE NOT NULL,
  "title" TEXT,
  "description" TEXT,
  "last_visit" INTEGER
);

CREATE TABLE IF NOT EXISTS "urls_meta" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "url_md5" VARCHAR(32) NOT NULL REFERENCES urls(url_md5),
  "indexed_at" INTEGER
);

CREATE TABLE IF NOT EXISTS "visits" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "url_md5" VARCHAR(32) NOT NULL REFERENCES urls(url_md5),
  "visit_time" INTEGER,
  "extractor_name" TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS visits_unique ON visits(url_md5, visit_time);
CREATE INDEX IF NOT EXISTS visits_url_md5 ON visits(url_md5);
`

// Open a connection to the database. Calling code should close the connection when done.
// @note It is assumed that the database is already initialized. Thus this may be less useful than `InitDB`
func OpenConnection(ctx context.Context, c *config.AppConfig) (*sql.DB, error) {
	dbPath := c.DBPath
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	return conn, err
}

// Initialize the database. Create tables and indexes
func InitDb(ctx context.Context, c *config.AppConfig) (*sql.DB, error) {
	conn, err := OpenConnection(ctx, c)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, initSql)

	return conn, err
}

func GetLatestTime(ctx context.Context, db *sql.DB, extractor types.Extractor) (*time.Time, error) {
	qry := `
SELECT
  visit_time
FROM
  visits
WHERE extractor_name = ?
ORDER BY
  visit_time DESC
LIMIT 1;
	`
	row := db.QueryRowContext(ctx, qry, extractor.GetName())
	if err := row.Err(); err != nil {
		return nil, err
	}

	var ts int64
	err := row.Scan(&ts)
	if err != nil {
		return nil, err
	}

	t := time.Unix(ts, 0)

	return &t, nil

}

func InsertUrl(ctx context.Context, db *sql.DB, row *types.UrlRow) error {
	const qry = `
		INSERT OR REPLACE INTO urls(url_md5, url, title, description, last_visit)
			VALUES(?, ?, ?, ?, ?);
	`
	var lastVisit int64
	if row.LastVisit != nil {
		lastVisit = row.LastVisit.Unix()
	}
	md5 := util.HashMd5String(row.Url)

	_, err := db.ExecContext(ctx, qry, md5, row.Url, row.Title, row.Description, lastVisit)
	return err
}

func InsertUrlMeta(ctx context.Context, db *sql.DB, row *types.UrlMetaRow) error {
	const qry = `
		INSERT OR REPLACE INTO urls_meta(url_md5, indexed_at)
			VALUES(?, ?);
	`
	md5 := util.HashMd5String(row.Url)

	_, err := db.ExecContext(ctx, qry, md5, row.IndexedAt)
	return err
}

func InsertVisit(ctx context.Context, db *sql.DB, row *types.VisitRow) error {
	const qry = `
		INSERT OR IGNORE INTO visits(url_md5, visit_time, extractor_name)
			VALUES(?, ?, ?);
	`
	md5 := util.HashMd5String(row.Url)

	_, err := db.ExecContext(ctx, qry, md5, row.Datetime.Unix(), row.ExtractorName)
	return err
}
