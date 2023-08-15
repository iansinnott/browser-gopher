package testutils

import (
	"database/sql"
	_ "embed"
	"sort"
	"strings"
	"testing"

	"github.com/iansinnott/browser-gopher/pkg/persistence"
	"github.com/pkg/errors"
)

// get an in-memory db connection. Don't forget to close your connection when done.
func GetTestDBConn(t *testing.T) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, errors.Wrap(err, "could not open test db")
	}

	entries, err := persistence.MigrationsDir.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	// make sure the migrations are sorted
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		// skip files that are not migrations
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filePath := "migrations/" + entry.Name()

		migration, err := persistence.MigrationsDir.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		_, err = conn.Exec(string(migration))
		if err != nil {
			return nil, err
		}
	}

	return conn, err
}
