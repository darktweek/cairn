package model

import "time"

// Collection permission levels (per share / effective access), ordered.
const (
	PermView   = "view"
	PermEdit   = "edit"
	PermManage = "manage"
)

// PermRank returns the ordered weight of a collection permission (higher = more).
func PermRank(p string) int {
	switch p {
	case PermView:
		return 1
	case PermEdit:
		return 2
	case PermManage:
		return 3
	default:
		return 0
	}
}

// PermAtLeast reports whether the held permission satisfies the required one.
func PermAtLeast(have, need string) bool {
	return PermRank(have) >= PermRank(need)
}

// Instance policy keys (stored in the settings table).
const (
	PolicyAdminManageAllCollections = "policy.admin_manage_all_collections"
	PolicyRestrictCollectionCreate  = "policy.restrict_collection_creation"
	PolicyRestrictCollectionDelete  = "policy.restrict_collection_deletion"
)

// CollectionShare is a per-user grant on a collection.
type CollectionShare struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Perm     string `json:"perm"`
}

// Collection is a shareable container of bookmarks. Each user owns exactly one
// personal collection (IsPersonal); all others are user-created and shareable.
type Collection struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	IsPersonal  bool      `json:"is_personal"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// OwnerUsername and Perm are populated for list/detail views; they are not
	// stored on the collections table.
	OwnerUsername string `json:"owner_username,omitempty"`
	Perm          string `json:"perm,omitempty"`
	BookmarkCount int    `json:"bookmark_count"`
	// Shared is true when the collection is part of a sharing relationship
	// (shared with me, or owned by me and shared with someone).
	Shared bool `json:"shared"`
}
