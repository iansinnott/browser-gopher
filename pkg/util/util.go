package util

import "time"

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
