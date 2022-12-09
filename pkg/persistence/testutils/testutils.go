package testutils

import (
	"database/sql"
	_ "embed"
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

	// Create the tables
	_, err = conn.Exec(persistence.InitSql)

	return conn, err
}
