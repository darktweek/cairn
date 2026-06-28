package model

import "time"

// Group is a named team of users; collections can be shared with a whole group.
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerID     string    `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember is a user's membership in a group.
type GroupMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"` // admin | member
}

// CollectionGroupShare is a per-group grant on a collection.
type CollectionGroupShare struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Perm      string `json:"perm"`
}
