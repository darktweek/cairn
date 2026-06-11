package repository

import "database/sql"

type Repositories struct {
	Users                UserRepository
	Sessions             SessionRepository
	Bookmarks            BookmarkRepository
	Tags                 TagRepository
	Wallpapers           WallpaperRepository
	TOTP                 TOTPRepository
	Audit                AuditRepository
	Invitations          InvitationRepository
	Settings             SettingsRepository
	PendingRegistrations PendingRegistrationRepository
}

func New(db *sql.DB) *Repositories {
	return &Repositories{
		Users:                newSQLiteUserRepo(db),
		Sessions:             newSQLiteSessionRepo(db),
		Bookmarks:            newSQLiteBookmarkRepo(db),
		Tags:                 newSQLiteTagRepo(db),
		Wallpapers:           newSQLiteWallpaperRepo(db),
		TOTP:                 newSQLiteTOTPRepo(db),
		Audit:                newSQLiteAuditRepo(db),
		Invitations:          newSQLiteInvitationRepo(db),
		Settings:             newSQLiteSettingsRepo(db),
		PendingRegistrations: newSQLitePendingRegRepo(db),
	}
}
