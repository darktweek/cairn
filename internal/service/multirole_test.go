package service

import (
	"context"
	"errors"
	"testing"

	"github.com/darktweek/cairn/internal/model"
)

func hasPerm(perms []string, p string) bool {
	for _, x := range perms {
		if x == p {
			return true
		}
	}
	return false
}

func TestMultipleRolesUnionPermissions(t *testing.T) {
	_, repos := newTestAuth(t)
	rbac := newRBACService(repos)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	if err := repos.Roles.AssignUserRole(ctx, owner.ID, model.RoleIDOwner, "admin"); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	ownerActor := loadActor(t, repos, owner.ID)

	rAudit, err := rbac.CreateRole(ctx, ownerActor, "auditor", []string{model.PermAuditView})
	if err != nil {
		t.Fatalf("create auditor: %v", err)
	}
	rGroups, err := rbac.CreateRole(ctx, ownerActor, "groupmgr", []string{model.PermGroupsManage})
	if err != nil {
		t.Fatalf("create groupmgr: %v", err)
	}

	target := makeUser(t, repos, "target", "target@example.com")
	if err := rbac.SetUserRoles(ctx, ownerActor, target.ID, []string{rAudit.ID, rGroups.ID}); err != nil {
		t.Fatalf("set roles: %v", err)
	}

	// Permissions are the union of both roles.
	perms, err := repos.Roles.PermissionsForUser(ctx, target.ID)
	if err != nil {
		t.Fatalf("perms: %v", err)
	}
	if !hasPerm(perms, model.PermAuditView) || !hasPerm(perms, model.PermGroupsManage) {
		t.Fatalf("union perms missing: %v", perms)
	}

	// The user reports two roles.
	roles, err := repos.Roles.RolesForUser(ctx, target.ID)
	if err != nil {
		t.Fatalf("roles for user: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("got %d roles, want 2", len(roles))
	}

	// Replacing with a single role drops the other permission.
	if err := rbac.SetUserRoles(ctx, ownerActor, target.ID, []string{rAudit.ID}); err != nil {
		t.Fatalf("reduce roles: %v", err)
	}
	perms, _ = repos.Roles.PermissionsForUser(ctx, target.ID)
	if hasPerm(perms, model.PermGroupsManage) {
		t.Fatalf("groups.manage should be gone after role removal: %v", perms)
	}
}

func TestSetUserRolesAntiEscalationAndLastOwner(t *testing.T) {
	_, repos := newTestAuth(t)
	rbac := newRBACService(repos)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	admin := makeUser(t, repos, "admin", "admin@example.com")
	_ = repos.Roles.AssignUserRole(ctx, owner.ID, model.RoleIDOwner, "admin")
	_ = repos.Roles.AssignUserRole(ctx, admin.ID, model.RoleIDAdmin, "admin")
	adminActor := loadActor(t, repos, admin.ID)
	ownerActor := loadActor(t, repos, owner.ID)

	// Admin cannot grant the owner role (anti-escalation).
	target := makeUser(t, repos, "t", "t@example.com")
	if err := rbac.SetUserRoles(ctx, adminActor, target.ID, []string{model.RoleIDOwner}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin granting owner: got %v, want ErrForbidden", err)
	}

	// Removing the owner role from the only owner is refused.
	if err := rbac.SetUserRoles(ctx, ownerActor, owner.ID, []string{model.RoleIDUser}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("demoting last owner: got %v, want ErrForbidden", err)
	}
}
