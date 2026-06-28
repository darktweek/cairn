package service

import (
	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/repository"
)

type Services struct {
	Auth       AuthService
	User       UserService
	Bookmark   BookmarkService
	Collection CollectionService
	Group      GroupService
	RBAC       RBACService
	Wallpaper  WallpaperService
	Admin      AdminService
	Email      EmailService
	Invitation InvitationService
	Settings   SettingsService
	OIDC       OIDCService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	settings := newSettingsService(repos, cfg)
	email := newEmailService(cfg, settings)
	auth := newAuthService(repos, cfg, settings, email)
	return &Services{
		Auth:       auth,
		User:       newUserService(repos, cfg),
		Bookmark:   newBookmarkService(repos, cfg, auth),
		Collection: newCollectionService(repos, email),
		Group:      newGroupService(repos),
		RBAC:       newRBACService(repos),
		Wallpaper:  newWallpaperService(repos, cfg, settings),
		Admin:      newAdminService(repos, cfg, email),
		Email:      email,
		Invitation: newInvitationService(repos, cfg, email),
		Settings:   settings,
		OIDC:       newOIDCService(cfg, settings),
	}
}
