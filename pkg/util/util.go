package util

import "time"

const SQLiteDateTime = "2006-01-02 15:04:05"

// Given a datetime string in the form "2022-01-14 06:41:48" parse it to time.Time
func ParseSQLiteDatetime(s string) (time.Time, error) {
	return time.Parse(SQLiteDateTime, s)
}
