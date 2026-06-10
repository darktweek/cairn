package service

import (
	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/repository"
)

type Services struct {
	Auth       AuthService
	User       UserService
	Bookmark   BookmarkService
	Wallpaper  WallpaperService
	Admin      AdminService
	Email      EmailService
	Invitation InvitationService
	Settings   SettingsService
	OIDC       OIDCService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	settings := newSettingsService(repos, cfg)
	auth := newAuthService(repos, cfg, settings)
	email := newEmailService(cfg, settings)
	return &Services{
		Auth:       auth,
		User:       newUserService(repos, cfg),
		Bookmark:   newBookmarkService(repos, cfg, auth),
		Wallpaper:  newWallpaperService(repos, cfg, settings),
		Admin:      newAdminService(repos, cfg),
		Email:      email,
		Invitation: newInvitationService(repos, cfg, email),
		Settings:   settings,
		OIDC:       newOIDCService(cfg, settings),
	}
}
