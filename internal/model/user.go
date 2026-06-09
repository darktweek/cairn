package model

import "time"

type User struct {
	ID               string
	Username         string
	Email            string
	Password         string // argon2id hash, never exposed in JSON
	Role             string // "user" | "admin"
	IsActive         bool
	WallpaperLimit   *int    // nil = use global config default
	SearchEngine     string  // "duckduckgo" | "google" | "brave" | "bing" | "kagi" | "custom"
	SearchEngineURL  *string // set only when SearchEngine == "custom"
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
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
