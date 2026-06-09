package model

import "time"

type Invitation struct {
	ID        string
	Email     string
	TokenHash string
	CreatedBy string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (i *Invitation) IsExpired() bool  { return time.Now().After(i.ExpiresAt) }
func (i *Invitation) IsUsed() bool     { return i.UsedAt != nil }
func (i *Invitation) IsPending() bool  { return !i.IsUsed() && !i.IsExpired() }
