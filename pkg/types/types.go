package types

import "time"

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
	GetAllUrls() ([]UrlRow, error)
	GetAllVisits() ([]VisitRow, error)
}
