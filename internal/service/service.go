package service

import (
	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/repository"
)

type Services struct {
	Auth      AuthService
	User      UserService
	Bookmark  BookmarkService
	Wallpaper WallpaperService
	Admin     AdminService
	Email     EmailService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	auth := newAuthService(repos, cfg)
	return &Services{
		Auth:      auth,
		User:      newUserService(repos, cfg),
		Bookmark:  newBookmarkService(repos, cfg, auth),
		Wallpaper: newWallpaperService(repos, cfg),
		Admin:     newAdminService(repos, cfg),
		Email:     newEmailService(cfg),
	}
}
