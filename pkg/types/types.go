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
	DocumentMd5 string
	UrlMd5      string
	StatusCode  int        // the HTTP status code returned during fetch
	AccessedAt  *time.Time // Nullable
	Body        *string    // Fulltext of the webpage as markdown
}

// Initially this was a URL row representation but it was later augmented with
// body, which is only available via join.
type UrlDbEntity struct {
	UrlMd5      string
	Url         string
	Title       *string
	Description *string
	LastVisit   *time.Time
	Body        *string
	BodyMd5     *string
}

type UrlDbSearchEntity struct {
	UrlMd5      string
	Url         string
	Title       *string
	Description *string
	LastVisit   *time.Time
	Match       *string
	MatchCount  *int
	SumRank     *float64
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

type SearchableEntity struct {
	Id          string     `json:"id"`
	Url         string     `json:"url"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	LastVisit   *time.Time `json:"last_visit"`
	Match       *string    `json:"match"`
	MatchCount  *int       `json:"match_count"`
	SumRank     *float64   `json:"sum_rank"`
}

func UrlDbEntityToSearchableEntity(x UrlDbEntity) SearchableEntity {
	return SearchableEntity{
		Id:          x.UrlMd5,
		Url:         x.Url,
		Title:       x.Title,
		Description: x.Description,
		LastVisit:   x.LastVisit,
	}
}

func UrlDbSearchEntityToSearchableEntity(x UrlDbSearchEntity) SearchableEntity {
	return SearchableEntity{
		Id:          x.UrlMd5,
		Url:         x.Url,
		Title:       x.Title,
		Description: x.Description,
		LastVisit:   x.LastVisit,
		Match:       x.Match,
		MatchCount:  x.MatchCount,
		SumRank:     x.SumRank,
	}
}
