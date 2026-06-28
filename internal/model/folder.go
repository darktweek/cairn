package model

import "time"

// Folder is a node in the folder tree of a single collection. ParentID nil = root.
type Folder struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	ParentID     *string   `json:"parent_id"`
	Name         string    `json:"name"`
	Sort         int       `json:"sort"`
	CreatedAt    time.Time `json:"created_at"`
}
