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
}
