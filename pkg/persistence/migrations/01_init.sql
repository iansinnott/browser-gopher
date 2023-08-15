CREATE TABLE IF NOT EXISTS "urls" (
  "url_md5" VARCHAR(32) PRIMARY KEY NOT NULL,
  "url" TEXT UNIQUE NOT NULL,
  "title" TEXT,
  "description" TEXT,
  "last_visit" INTEGER
);

CREATE TABLE IF NOT EXISTS "urls_meta" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "url_md5" VARCHAR(32) UNIQUE NOT NULL REFERENCES urls(url_md5),
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

CREATE TABLE IF NOT EXISTS "documents" (
  "document_md5" VARCHAR(32) PRIMARY KEY NOT NULL,
	"body" TEXT,
	"status_code" INTEGER,
  "accessed_at" INTEGER
);

CREATE TABLE IF NOT EXISTS "url_document_edges" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "url_md5" VARCHAR(32) UNIQUE NOT NULL REFERENCES urls(url_md5),
  "document_md5" VARCHAR(32) NOT NULL REFERENCES documents(document_md5)
);

CREATE VIEW IF NOT EXISTS "searchable_data" AS
SELECT
	urls.rowid as url_rowid,
	urls.url_md5,
	urls.url,
	urls.title,
	urls.description,
	documents.document_md5,
	documents.body
FROM
	urls
	LEFT OUTER JOIN url_document_edges ON urls.url_md5 = url_document_edges.url_md5
	LEFT OUTER JOIN documents ON url_document_edges.document_md5 = documents.document_md5;