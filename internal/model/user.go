package model

import "time"

type User struct {
	ID              string
	Username        string
	Email           string
	Password        string `json:"-"` // argon2id hash, never exposed in JSON
	Role            string // legacy coarse role column: "user" | "admin"
	RoleID          string // primary role id (denormalised, for single-role display)
	RoleName        string // primary role display name
	Roles           []Role // all roles held by the user (multi-role); loaded when needed
	Permissions     []string // union of permissions across all roles, loaded on auth
	IsActive        bool
	WallpaperLimit   *int   // nil = use global config default
	UploadSizeLimit  *int64 // max bytes for a single file upload; nil = use global CAIRN_MAX_UPLOAD_SIZE
	StorageQuota     *int64 // max total media bytes; nil = use global CAIRN_STORAGE_QUOTA
	SearchEngine    string  // "duckduckgo" | "google" | "brave" | "bing" | "kagi" | "custom"
	SearchEngineURL *string // set only when SearchEngine == "custom"
	Locale          string  // "fr" | "en" — interface language chosen by the user
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Session struct {
	ID            string
	UserID        string
	TokenHash     string
	UserAgent     string
	IP            string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	IsBookmarklet bool
}

type TOTPSecret struct {
	UserID     string
	Secret     string // AES-GCM encrypted
	IsVerified bool
	CreatedAt  time.Time
}
