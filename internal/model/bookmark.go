package model

import "time"

type Bookmark struct {
	ID           string    `json:"id"`
	UserID       string    `json:"-"`
	CollectionID string    `json:"collection_id"`
	FolderID     *string   `json:"folder_id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Folder       *string   `json:"folder"` // deprecated: legacy folder string, superseded by FolderID
	Sort         int       `json:"sort"`
	Tags         []Tag     `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// AddedByUsername is populated for shared-collection list views (the bookmark author).
	AddedByUsername string `json:"added_by_username,omitempty"`
}

type Tag struct {
	ID     string `json:"id"`
	UserID string `json:"-"`
	Name   string `json:"name"`
}
