package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

type RBACService interface {
	Catalog() []string
	ListRoles(ctx context.Context) ([]*model.Role, error)
	CreateRole(ctx context.Context, actor *model.User, name string, perms []string) (*model.Role, error)
	UpdateRole(ctx context.Context, actor *model.User, id, name string, perms []string) error
	DeleteRole(ctx context.Context, id string) error
	AssignRole(ctx context.Context, actor *model.User, userID, roleID string) error
	SetUserRoles(ctx context.Context, actor *model.User, userID string, roleIDs []string) error
	RolesForUser(ctx context.Context, userID string) ([]model.Role, error)
	RolesForUsers(ctx context.Context, userIDs []string) (map[string][]model.Role, error)
}

type rbacService struct {
	repos *repository.Repositories
}

func newRBACService(repos *repository.Repositories) RBACService {
	return &rbacService{repos: repos}
}

func (s *rbacService) Catalog() []string { return model.AllPermissions }

func (s *rbacService) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return s.repos.Roles.ListAll(ctx)
}

// guardGrantable enforces the Bitwarden rule: an actor may only grant permissions
// they themselves hold.
func guardGrantable(actor *model.User, perms []string) error {
	for _, p := range perms {
		if !model.IsValidPermission(p) {
			return fmt.Errorf("%w: unknown permission %q", ErrInvalidInput, p)
		}
		if !actor.Can(p) {
			return fmt.Errorf("%w: cannot grant a permission you do not hold (%s)", ErrForbidden, p)
		}
	}
	return nil
}

func (s *rbacService) CreateRole(ctx context.Context, actor *model.User, name string, perms []string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: role name required", ErrInvalidInput)
	}
	if err := guardGrantable(actor, perms); err != nil {
		return nil, err
	}
	now := time.Now()
	role := &model.Role{
		ID:          ulid.Make().String(),
		Name:        name,
		IsSystem:    false,
		Permissions: perms,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repos.Roles.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConflict, err.Error())
	}
	return role, nil
}

func (s *rbacService) UpdateRole(ctx context.Context, actor *model.User, id, name string, perms []string) error {
	role, err := s.repos.Roles.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if role.IsSystem {
		return fmt.Errorf("%w: system roles cannot be modified", ErrForbidden)
	}
	if err := guardGrantable(actor, perms); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: role name required", ErrInvalidInput)
	}
	role.Name = name
	role.Permissions = perms
	return s.repos.Roles.Update(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, id string) error {
	role, err := s.repos.Roles.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if role.IsSystem {
		return fmt.Errorf("%w: system roles cannot be deleted", ErrForbidden)
	}
	n, err := s.repos.Roles.CountUsersWithRole(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: role is assigned to %d user(s)", ErrConflict, n)
	}
	return s.repos.Roles.Delete(ctx, id)
}

// AssignRole sets a single role (convenience over SetUserRoles).
func (s *rbacService) AssignRole(ctx context.Context, actor *model.User, userID, roleID string) error {
	return s.SetUserRoles(ctx, actor, userID, []string{roleID})
}

// SetUserRoles replaces a user's full set of roles.
func (s *rbacService) SetUserRoles(ctx context.Context, actor *model.User, userID string, roleIDs []string) error {
	target, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	// Privilege guard: an actor may not modify a user who holds a permission the
	// actor lacks (you cannot tamper with someone more privileged than yourself).
	// This stops a plain users.manage admin from stripping an owner/admin's roles.
	targetPerms, err := s.repos.Roles.PermissionsForUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, p := range targetPerms {
		if !actor.Can(p) {
			return fmt.Errorf("%w: cannot modify a more privileged user", ErrForbidden)
		}
	}

	// Load the requested roles, dedupe, and gather their permission union.
	seen := map[string]bool{}
	var roles []*model.Role
	var union []string
	hasOwner := false
	for _, rid := range roleIDs {
		if rid == "" || seen[rid] {
			continue
		}
		seen[rid] = true
		role, err := s.repos.Roles.GetByID(ctx, rid)
		if err != nil {
			return fmt.Errorf("%w: unknown role", ErrInvalidInput)
		}
		roles = append(roles, role)
		union = append(union, role.Permissions...)
		if role.ID == model.RoleIDOwner {
			hasOwner = true
		}
	}

	// Anti-escalation: the actor may only grant permissions they hold.
	if err := guardGrantable(actor, union); err != nil {
		return err
	}

	// Never strip the instance of its last owner.
	targetWasOwner := false
	for _, rl := range mustRoles(ctx, s, userID) {
		if rl.ID == model.RoleIDOwner {
			targetWasOwner = true
		}
	}
	if targetWasOwner && !hasOwner {
		owners, err := s.repos.Roles.CountUsersWithRole(ctx, model.RoleIDOwner)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return fmt.Errorf("%w: cannot demote the last owner", ErrForbidden)
		}
	}

	ids := make([]string, 0, len(roles))
	for _, rl := range roles {
		ids = append(ids, rl.ID)
	}
	if err := s.repos.Roles.SetUserRoles(ctx, userID, ids, primaryRoleID(roles), legacyRoleForSet(roles)); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &actor.ID,
		Action:    "role_assigned",
		Metadata:  map[string]any{"target": userID, "roles": ids},
		CreatedAt: time.Now(),
	})
	_ = target
	return nil
}

func mustRoles(ctx context.Context, s *rbacService, userID string) []model.Role {
	roles, _ := s.repos.Roles.RolesForUser(ctx, userID)
	return roles
}

func (s *rbacService) RolesForUser(ctx context.Context, userID string) ([]model.Role, error) {
	return s.repos.Roles.RolesForUser(ctx, userID)
}

func (s *rbacService) RolesForUsers(ctx context.Context, userIDs []string) (map[string][]model.Role, error) {
	return s.repos.Roles.RolesForUsers(ctx, userIDs)
}

// primaryRoleID picks the representative role for single-role display:
// owner > admin > first custom > user.
func primaryRoleID(roles []*model.Role) string {
	if len(roles) == 0 {
		return ""
	}
	var admin, custom, user string
	for _, r := range roles {
		switch r.ID {
		case model.RoleIDOwner:
			return model.RoleIDOwner
		case model.RoleIDAdmin:
			admin = r.ID
		case model.RoleIDUser:
			user = r.ID
		default:
			if custom == "" {
				custom = r.ID
			}
		}
	}
	if admin != "" {
		return admin
	}
	if custom != "" {
		return custom
	}
	return user
}

// legacyRoleForSet returns "admin" if any role grants an admin-area permission,
// keeping users.role (CHECK + admin-count guard) consistent.
func legacyRoleForSet(roles []*model.Role) string {
	for _, role := range roles {
		for _, p := range role.Permissions {
			if p == model.PermUsersManage || p == model.PermRolesManage || p == model.PermSettingsManage {
				return "admin"
			}
		}
	}
	return "user"
}
