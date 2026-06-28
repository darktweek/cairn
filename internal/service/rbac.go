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

func (s *rbacService) AssignRole(ctx context.Context, actor *model.User, userID, roleID string) error {
	role, err := s.repos.Roles.GetByID(ctx, roleID)
	if err != nil {
		return ErrNotFound
	}
	// The actor may only assign a role whose permissions they could grant.
	if err := guardGrantable(actor, role.Permissions); err != nil {
		return err
	}

	target, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	// Never strip the instance of its last owner.
	if target.RoleID == model.RoleIDOwner && roleID != model.RoleIDOwner {
		owners, err := s.repos.Roles.CountUsersWithRole(ctx, model.RoleIDOwner)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return fmt.Errorf("%w: cannot demote the last owner", ErrForbidden)
		}
	}

	if err := s.repos.Roles.AssignUserRole(ctx, userID, roleID, legacyRoleFor(role)); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &actor.ID,
		Action:    "role_assigned",
		Metadata:  map[string]any{"target": userID, "role": role.Name},
		CreatedAt: time.Now(),
	})
	return nil
}

// legacyRoleFor maps a role to the coarse user/admin value kept in users.role
// (so the old CHECK constraint and admin-count guard keep working).
func legacyRoleFor(role *model.Role) string {
	for _, p := range role.Permissions {
		if p == model.PermUsersManage || p == model.PermRolesManage || p == model.PermSettingsManage {
			return "admin"
		}
	}
	return "user"
}
