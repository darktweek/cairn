package model

import "time"

// Instance-level permissions (Bitwarden-style "custom role" capabilities).
// These gate real features and are frozen in code; roles bundle subsets of them.
const (
	PermAuditView            = "audit.view"
	PermBookmarksIO          = "bookmarks.import_export"
	PermCollectionsCreate    = "collections.create"
	PermCollectionsManageAll = "collections.manage_all" // gated by the admin-override policy
	PermCollectionsDeleteAny = "collections.delete_any"
	PermGroupsManage         = "groups.manage"
	PermUsersManage          = "users.manage"
	PermSettingsManage       = "settings.manage"
	PermRolesManage          = "roles.manage"
)

// AllPermissions is the catalog exposed to the admin UI for building roles.
var AllPermissions = []string{
	PermAuditView,
	PermBookmarksIO,
	PermCollectionsCreate,
	PermCollectionsManageAll,
	PermCollectionsDeleteAny,
	PermGroupsManage,
	PermUsersManage,
	PermSettingsManage,
	PermRolesManage,
}

// Stable IDs of the seeded system roles (see migration 019).
const (
	RoleIDOwner = "role_owner"
	RoleIDAdmin = "role_admin"
	RoleIDUser  = "role_user"
)

// Role is a named bundle of instance permissions. System roles cannot be deleted.
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsValidPermission reports whether p is a known permission key.
func IsValidPermission(p string) bool {
	for _, k := range AllPermissions {
		if k == p {
			return true
		}
	}
	return false
}

// Can reports whether the user's loaded permission set includes perm.
func (u *User) Can(perm string) bool {
	for _, p := range u.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}
