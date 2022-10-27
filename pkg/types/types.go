package types

import (
	"context"
	"database/sql"
	"time"
)

type UrlRow struct {
	Url         string
	Title       *string    // Nullable
	Description *string    // Nullable
	LastVisit   *time.Time // Nullable
}

// Meta information about the URL
type UrlMetaRow struct {
	Url       string
	IndexedAt *time.Time // Nullable
}

// DocumentRow represents a full-text document. The HTML version of a web page.
// However, the HTML body is not stored (for now). The page will be distilled to
// plain text. A markdown version will be stored on disk, again, for now.
type DocumentRow struct {
	DocumentMd5  string
	UrlMd5       string
	MarkdownPath string     // path on disk to stored markdown
	StatusCode   int        // the HTTP status code returned during fetch
	AccessedAt   *time.Time // Nullable
}

// The URL as represented in the db.
type UrlDbEntity struct {
	UrlMd5      string
	Url         string
	Title       *string    // Nullable
	Description *string    // Nullable
	LastVisit   *time.Time // Nullable
}

type VisitRow struct {
	Url      string
	Datetime time.Time
	// The data extractor that created this visit. Not present on URls since URLs
	// are often visited in multiple browsers.
	ExtractorName string
}

type Extractor interface {
	GetName() string
	GetDBPath() string
	SetDBPath(string)
	GetAllUrlsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]UrlRow, error)
	GetAllVisitsSince(ctx context.Context, conn *sql.DB, since time.Time) ([]VisitRow, error)

	// Verify that the passed db can actually be connected to. In the case of
	// sqlite, it's not uncommon for a db to be locked. The Open call will work
	// but the db cannot be read.
	VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error)
}
