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

CREATE VIRTUAL TABLE IF NOT EXISTS "search" USING fts5(
	url_md5 UNINDEXED,
	body_md5 UNINDEXED,
	url,
	title,
	description,
	body,
	tokenize = 'porter unicode61 remove_diacritics 2'
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

-- @note Our code does many INSERT OR REPLACE into the URLs table, which ends up
-- firing this trigger _FOR A REPLACE_ operation. That means this trigger is
-- similar to the AFTER UPDATE trigger except it does not come with an OLD value
-- for us to use. Populating from the VIEW is so that an OR REPLACE won't blow
-- away existing document body data, if present.
CREATE TRIGGER IF NOT EXISTS search_insert_urls AFTER INSERT ON urls
BEGIN
	DELETE FROM search WHERE url_md5 = NEW.url_md5;
	INSERT INTO search (rowid, url_md5, url, title, description, body_md5, body) 
		SELECT url_rowid, url_md5, url, title, description, document_md5, body
		FROM searchable_data WHERE url_md5 = NEW.url_md5;
END;

-- @note It's important NOT to use the values('delete') form of index deletion
-- since we're not using a content table. Since this is a trigger, it will
-- silently fail.
-- Ex (of what not to do, unless using a contentless search index): 
--
-- 			INSERT INTO search (search, ...) VALUES ('delete', ...);
--
CREATE TRIGGER IF NOT EXISTS search_delete_urls AFTER DELETE ON urls
BEGIN
  DELETE FROM search WHERE rowid = OLD.rowid;
END;

CREATE TRIGGER IF NOT EXISTS search_update_urls AFTER UPDATE ON urls
BEGIN
  DELETE FROM search WHERE rowid = OLD.rowid;
	INSERT INTO search (rowid, url_md5, url, title, description, body_md5, body) 
		SELECT url_rowid, url_md5, url, title, description, document_md5, body
		FROM searchable_data WHERE url_md5 = NEW.url_md5;
END;

CREATE TRIGGER IF NOT EXISTS search_insert_document AFTER INSERT ON url_document_edges
BEGIN
	DELETE FROM search WHERE url_md5 = NEW.url_md5;
	INSERT INTO search (rowid, url_md5, url, title, description, body_md5, body) 
		SELECT url_rowid, url_md5, url, title, description, document_md5, body
		FROM searchable_data WHERE url_md5 = NEW.url_md5;
END;

CREATE TRIGGER IF NOT EXISTS search_delete_document AFTER DELETE ON url_document_edges
BEGIN
	UPDATE "search" SET body = NULL, body_md5 = NULL
		WHERE url_md5 = OLD.url_md5 AND body_md5 = OLD.document_md5;
END;

CREATE TRIGGER IF NOT EXISTS search_update_document AFTER UPDATE ON url_document_edges
BEGIN
	DELETE FROM search WHERE url_md5 = NEW.url_md5;
	INSERT INTO search (rowid, url_md5, url, title, description, body_md5, body) 
		SELECT url_rowid, url_md5, url, title, description, document_md5, body
		FROM searchable_data WHERE url_md5 = NEW.url_md5;
END;