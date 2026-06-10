package model

import "time"

type Wallpaper struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Filename  string    `json:"filename"`
	IsPinned  bool      `json:"is_pinned"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"created_at"`
}
