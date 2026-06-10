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
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	auth  := newAuthService(repos, cfg)
	email := newEmailService(cfg)
	return &Services{
		Auth:       auth,
		User:       newUserService(repos, cfg),
		Bookmark:   newBookmarkService(repos, cfg, auth),
		Wallpaper:  newWallpaperService(repos, cfg),
		Admin:      newAdminService(repos, cfg),
		Email:      email,
		Invitation: newInvitationService(repos, cfg, email),
		Settings:   newSettingsService(repos, cfg),
	}
}
