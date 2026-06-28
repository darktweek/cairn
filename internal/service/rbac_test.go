package service

import (
	"context"
	"errors"
	"testing"

	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
)

// loadActor builds an authenticated *model.User with its permission set loaded.
func loadActor(t *testing.T, repos *repository.Repositories, userID string) *model.User {
	t.Helper()
	u, err := repos.Users.GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get actor: %v", err)
	}
	perms, err := repos.Roles.PermissionsForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("perms: %v", err)
	}
	u.Permissions = perms
	return u
}

func TestRoleResolutionAndAntiEscalation(t *testing.T) {
	_, repos := newTestAuth(t)
	rbac := newRBACService(repos)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	admin := makeUser(t, repos, "admin", "admin@example.com")
	if err := repos.Roles.AssignUserRole(ctx, owner.ID, model.RoleIDOwner, "admin"); err != nil {
		t.Fatalf("assign owner: %v", err)
	}
	if err := repos.Roles.AssignUserRole(ctx, admin.ID, model.RoleIDAdmin, "admin"); err != nil {
		t.Fatalf("assign admin: %v", err)
	}

	// Permission resolution: owner has roles.manage, admin does not.
	ownerActor := loadActor(t, repos, owner.ID)
	adminActor := loadActor(t, repos, admin.ID)
	if !ownerActor.Can(model.PermRolesManage) {
		t.Fatal("owner should hold roles.manage")
	}
	if adminActor.Can(model.PermRolesManage) {
		t.Fatal("admin must not hold roles.manage")
	}
	if !adminActor.Can(model.PermUsersManage) {
		t.Fatal("admin should hold users.manage")
	}

	// Anti-escalation: admin cannot create a role granting a permission they lack.
	if _, err := rbac.CreateRole(ctx, adminActor, "escalator", []string{model.PermRolesManage}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin creating roles.manage role: got %v, want ErrForbidden", err)
	}
	// But can create a role within their own powers.
	if _, err := rbac.CreateRole(ctx, adminActor, "auditor", []string{model.PermAuditView}); err != nil {
		t.Fatalf("admin creating auditor role: %v", err)
	}

	// Anti-escalation on assignment: admin cannot grant the owner role.
	if err := rbac.AssignRole(ctx, adminActor, admin.ID, model.RoleIDOwner); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin assigning owner: got %v, want ErrForbidden", err)
	}
	// Owner can.
	target := makeUser(t, repos, "target", "target@example.com")
	if err := rbac.AssignRole(ctx, ownerActor, target.ID, model.RoleIDAdmin); err != nil {
		t.Fatalf("owner assigning admin: %v", err)
	}

	// Last-owner protection: demoting the only owner is refused.
	if err := rbac.AssignRole(ctx, ownerActor, owner.ID, model.RoleIDUser); !errors.Is(err, ErrForbidden) {
		t.Fatalf("demoting last owner: got %v, want ErrForbidden", err)
	}
}
