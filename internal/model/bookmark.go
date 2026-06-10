package model

import "time"

type Bookmark struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Folder    *string   `json:"folder"`
	Sort      int       `json:"sort"`
	Tags      []Tag     `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	ID     string `json:"id"`
	UserID string `json:"-"`
	Name   string `json:"name"`
}
