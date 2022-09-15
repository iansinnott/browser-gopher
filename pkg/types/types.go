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
