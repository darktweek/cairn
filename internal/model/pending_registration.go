package model

import "time"

type PendingRegistration struct {
	ID          string
	Username    string
	Email       string
	TokenHash   string
	TOTPSecret  string // encrypted AES-GCM, same scheme as totp_secrets.secret
	ExpiresAt   time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func (p *PendingRegistration) IsExpired() bool   { return time.Now().After(p.ExpiresAt) }
func (p *PendingRegistration) IsCompleted() bool { return p.CompletedAt != nil }
