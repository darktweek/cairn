package model

import "time"

type Bookmark struct {
	ID        string
	UserID    string
	URL       string
	Title     string
	Folder    *string // nil = no folder, "Dev/Go" = materialized path
	Sort      int
	Tags      []Tag
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID     string
	UserID string
	Name   string
}
