package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

type InvitationService interface {
	// Settings
	IsOpenRegistration(ctx context.Context) (bool, error)
	SetOpenRegistration(ctx context.Context, open bool) error

	// Invitations
	Create(ctx context.Context, adminID, email string) (*model.Invitation, string, error)
	Validate(ctx context.Context, token string) (*model.Invitation, error)
	Consume(ctx context.Context, token string) (*model.Invitation, error)
	List(ctx context.Context) ([]*model.Invitation, error)
	Revoke(ctx context.Context, adminID, id string) error
	Resend(ctx context.Context, id, adminID string) (string, error)
}

type invitationService struct {
	repos *repository.Repositories
	cfg   *config.Config
	email EmailService
}

func newInvitationService(repos *repository.Repositories, cfg *config.Config, email EmailService) InvitationService {
	return &invitationService{repos: repos, cfg: cfg, email: email}
}

func (s *invitationService) IsOpenRegistration(ctx context.Context) (bool, error) {
	v, err := s.repos.Settings.Get(ctx, "open_registration")
	if err != nil {
		return s.cfg.OpenRegistration, nil
	}
	return v == "true", nil
}

func (s *invitationService) SetOpenRegistration(ctx context.Context, open bool) error {
	v := "false"
	if open {
		v = "true"
	}
	return s.repos.Settings.Set(ctx, "open_registration", v)
}

func (s *invitationService) Create(ctx context.Context, adminID, email string) (*model.Invitation, string, error) {
	if email == "" {
		return nil, "", fmt.Errorf("%w: email required", ErrInvalidInput)
	}

	raw, hash, err := generateInviteToken()
	if err != nil {
		return nil, "", err
	}

	inv := &model.Invitation{
		ID:        ulid.Make().String(),
		Email:     email,
		TokenHash: hash,
		CreatedBy: adminID,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.InviteLifetime) * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := s.repos.Invitations.Create(ctx, inv); err != nil {
		return nil, "", err
	}

	link := fmt.Sprintf("%s/?invite=%s", s.cfg.BaseURL, raw)
	if err := s.email.SendInvitation(ctx, email, link, inv.ExpiresAt); err != nil {
		// non-fatal: admin can resend
		_ = err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &adminID,
		Action:    "invitation_sent",
		Metadata:  map[string]any{"email": email},
		CreatedAt: time.Now(),
	})

	return inv, raw, nil
}

func (s *invitationService) Validate(ctx context.Context, token string) (*model.Invitation, error) {
	hash := hashToken(token)
	inv, err := s.repos.Invitations.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrNotFound
	}
	if inv.IsUsed() {
		return nil, fmt.Errorf("%w: invitation already used", ErrInvalidInput)
	}
	if inv.IsExpired() {
		return nil, fmt.Errorf("%w: invitation expired", ErrInvalidInput)
	}
	return inv, nil
}

func (s *invitationService) Consume(ctx context.Context, token string) (*model.Invitation, error) {
	inv, err := s.Validate(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Invitations.MarkUsed(ctx, inv.ID, time.Now()); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *invitationService) List(ctx context.Context) ([]*model.Invitation, error) {
	return s.repos.Invitations.List(ctx)
}

func (s *invitationService) Revoke(ctx context.Context, adminID, id string) error {
	if err := s.repos.Invitations.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &adminID,
		Action:    "invitation_revoked",
		Metadata:  map[string]any{"invitation_id": id},
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *invitationService) Resend(ctx context.Context, id, adminID string) (string, error) {
	inv, err := s.repos.Invitations.GetByID(ctx, id)
	if err != nil {
		return "", ErrNotFound
	}
	if inv.IsUsed() {
		return "", fmt.Errorf("%w: invitation already used", ErrInvalidInput)
	}

	// Generate a fresh token and extend expiry.
	raw, hash, err := generateInviteToken()
	if err != nil {
		return "", err
	}
	inv.TokenHash = hash
	inv.ExpiresAt = time.Now().Add(time.Duration(s.cfg.InviteLifetime) * time.Hour)

	// Delete + recreate to update token_hash (simpler than UPDATE with UNIQUE).
	if err := s.repos.Invitations.Delete(ctx, id); err != nil {
		return "", err
	}
	inv.CreatedBy = adminID
	inv.CreatedAt = time.Now()
	if err := s.repos.Invitations.Create(ctx, inv); err != nil {
		return "", err
	}

	link := fmt.Sprintf("%s/?invite=%s", s.cfg.BaseURL, raw)
	if err := s.email.SendInvitation(ctx, inv.Email, link, inv.ExpiresAt); err != nil {
		_ = err
	}

	return raw, nil
}

func generateInviteToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = hashToken(raw)
	return
}
