package repository

import "github.com/oklog/ulid/v2"

// newID returns a fresh ULID string, matching the IDs minted by the service layer.
func newID() string {
	return ulid.Make().String()
}

// nullStr maps an empty string to SQL NULL, otherwise to the string value.
func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// boolInt maps a bool to the 0/1 integer SQLite stores for boolean columns.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
