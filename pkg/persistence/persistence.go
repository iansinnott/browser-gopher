package persistence

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/iansinnott/browser-gopher/pkg/types"
	"github.com/iansinnott/browser-gopher/pkg/util"
)

const initSql = `
CREATE TABLE IF NOT EXISTS "urls" (
  "url_md5" VARCHAR(32) PRIMARY KEY NOT NULL,
  "url" TEXT UNIQUE NOT NULL,
  "title" TEXT,
  "last_visit" INTEGER
);

CREATE TABLE IF NOT EXISTS "visits" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "url_md5" VARCHAR(32) NOT NULL REFERENCES urls(url_md5),
  "visit_time" INTEGER,
  "extractor_name" TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS visits_unique ON visits(extractor_name, url_md5, visit_time);
CREATE INDEX IF NOT EXISTS visits_url_md5 ON visits(url_md5);
`

func InitDB(ctx context.Context, c *config.AppConfig) (*sql.DB, error) {
	dbPath := c.DBPath
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(ctx, initSql)

	return conn, err
}

func InsertURL(ctx context.Context, db *sql.DB, row *types.UrlRow) error {
	const qry = `
		INSERT OR REPLACE INTO urls(url_md5, url, title, last_visit)
			VALUES(?, ?, ?, ?);
	`
	var lastVisit int64
	if row.LastVisit != nil {
		lastVisit = row.LastVisit.Unix()
	}
	md5 := util.HashMd5String(row.Url)

	_, err := db.ExecContext(ctx, qry, md5, row.Url, row.Title, lastVisit)
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
