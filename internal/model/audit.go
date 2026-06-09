package model

import "time"

type AuditEntry struct {
	ID        string
	UserID    *string
	Action    string
	IP        string
	UserAgent string
	Metadata  map[string]any // JSON-serialized in DB
	CreatedAt time.Time
}

type AdminStats struct {
	TotalUsers      int
	ActiveUsers     int
	TotalBookmarks  int
	TotalWallpapers int
	DBSizeBytes     int64
}
