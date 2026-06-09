package model

import "time"

type Wallpaper struct {
	ID        string
	UserID    string
	Filename  string
	IsPinned  bool
	Sort      int
	CreatedAt time.Time
}
