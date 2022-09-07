package types

import (
	"context"
	"database/sql"
	"time"
)

type UrlRow struct {
	Url         string
	Title       *string // Nullable
	Description *string // Nullable
}

type VisitRow struct {
	Url      string
	Datetime time.Time
}

type Extractor interface {
	GetName() string
	GetDBPath() string
	GetAllUrls(ctx context.Context, conn *sql.DB) ([]UrlRow, error)
	GetAllVisits(ctx context.Context, conn *sql.DB) ([]VisitRow, error)

	// Verify that the passed db can actually be connected to. In the case of
	// sqlite, it's not uncommon for a db to be locked. The Open call will work
	// but the db cannot be read.
	VerifyConnection(ctx context.Context, conn *sql.DB) (bool, error)
}
