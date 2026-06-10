package service

import (
	"context"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/repository"
)

// SettingsService centralizes runtime-configurable settings. Each setting can
// come from three sources, in priority order:
//
//	env     — hardcoded in the compose / environment, locked (read-only in admin)
//	db      — set by an admin through the panel
//	default — built-in fallback
//
// The "locked" flag tells the frontend whether the admin may edit a value.
type SettingsService interface {
	MenuBang(ctx context.Context) string
	SetMenuBang(ctx context.Context, bang string) error
	MenuBangLocked() bool
}

const defaultMenuBang = "!menu"

type settingsService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func newSettingsService(repos *repository.Repositories, cfg *config.Config) SettingsService {
	return &settingsService{repos: repos, cfg: cfg}
}

func (s *settingsService) MenuBang(ctx context.Context) string {
	if s.cfg.MenuBang != "" {
		return s.cfg.MenuBang // env-locked
	}
	if v, err := s.repos.Invitations.GetSetting(ctx, "menu_bang"); err == nil && v != "" {
		return v
	}
	return defaultMenuBang
}

func (s *settingsService) SetMenuBang(ctx context.Context, bang string) error {
	if s.MenuBangLocked() {
		return ErrForbidden
	}
	if bang == "" || bang[0] != '!' || len(bang) < 2 {
		return ErrInvalidInput
	}
	return s.repos.Invitations.SetSetting(ctx, "menu_bang", bang)
}

func (s *settingsService) MenuBangLocked() bool { return s.cfg.MenuBang != "" }
