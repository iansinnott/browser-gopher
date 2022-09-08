package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const SQLiteDateTime = "2006-01-02 15:04:05"

// Given a datetime string in the form "2022-01-14 06:41:48" parse it to time.Time
//
// @note Rather than parse timestamps we can also pull timestamps out of the db.
// Here's an example for Chrome:
//
//     strftime("%s", visit_time / 1e6 + strftime ('%s', '1601-01-01'), 'unixepoch') AS `timestamp`,
//
// Might be a better approach, but for now I like seeing the extracted time
// visually for debugging.
func ParseSQLiteDatetime(s string) (time.Time, error) {
	return time.Parse(SQLiteDateTime, s)
}

// A quick helper for parsing iso time because I find it hard to remember the const name
func ParseISODatetime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// Expand tilde in path strings
func Expanduser(path string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("could not get user home", err)
		os.Exit(1)
	}

	return strings.Replace(path, "~", userHome, 1)
}
