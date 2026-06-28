package repository

import "database/sql"

type Repositories struct {
	Users                UserRepository
	Roles                RoleRepository
	Sessions             SessionRepository
	Bookmarks            BookmarkRepository
	Collections          CollectionRepository
	Folders              FolderRepository
	Groups               GroupRepository
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
		Roles:                newSQLiteRoleRepo(db),
		Sessions:             newSQLiteSessionRepo(db),
		Bookmarks:            newSQLiteBookmarkRepo(db),
		Collections:          newSQLiteCollectionRepo(db),
		Folders:              newSQLiteFolderRepo(db),
		Groups:               newSQLiteGroupRepo(db),
		Tags:                 newSQLiteTagRepo(db),
		Wallpapers:           newSQLiteWallpaperRepo(db),
		TOTP:                 newSQLiteTOTPRepo(db),
		Audit:                newSQLiteAuditRepo(db),
		Invitations:          newSQLiteInvitationRepo(db),
		Settings:             newSQLiteSettingsRepo(db),
		PendingRegistrations: newSQLitePendingRegRepo(db),
	}
}
