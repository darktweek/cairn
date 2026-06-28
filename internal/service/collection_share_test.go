package service

import (
	"context"
	"errors"
	"testing"

	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
)

func TestCollectionSharingGrantsScopedAccess(t *testing.T) {
	auth, repos := newTestAuth(t)
	bm := newBookmarkService(repos, nil, auth)
	cols := newCollectionService(repos, nil)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	guest := makeUser(t, repos, "guest", "guest@example.com")

	col, err := cols.Create(ctx, owner.ID, CollectionInput{Name: "Shared"})
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if _, err := bm.Create(ctx, owner.ID, BookmarkInput{URL: "https://a.com", Title: "A", CollectionID: col.ID}); err != nil {
		t.Fatalf("owner add bookmark: %v", err)
	}

	// No access yet → hidden.
	if _, _, err := bm.ListInCollection(ctx, guest.ID, col.ID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("guest before share: got %v, want ErrNotFound", err)
	}

	// Share as viewer → can read, cannot write.
	if err := cols.SetShare(ctx, owner.ID, col.ID, guest.ID, model.PermView); err != nil {
		t.Fatalf("share view: %v", err)
	}
	if _, total, err := bm.ListInCollection(ctx, guest.ID, col.ID, repository.BookmarkFilter{Limit: 10}); err != nil || total != 1 {
		t.Fatalf("guest view list: err=%v total=%d", err, total)
	}
	if _, err := bm.Create(ctx, guest.ID, BookmarkInput{URL: "https://b.com", Title: "B", CollectionID: col.ID}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer write: got %v, want ErrForbidden", err)
	}

	// Upgrade to editor → can write.
	if err := cols.SetShare(ctx, owner.ID, col.ID, guest.ID, model.PermEdit); err != nil {
		t.Fatalf("share edit: %v", err)
	}
	if _, err := bm.Create(ctx, guest.ID, BookmarkInput{URL: "https://b.com", Title: "B", CollectionID: col.ID}); err != nil {
		t.Fatalf("editor write: %v", err)
	}

	// A viewer/editor cannot manage shares (manage required).
	if err := cols.SetShare(ctx, guest.ID, col.ID, owner.ID, model.PermView); !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor managing shares: got %v, want ErrForbidden", err)
	}

	// Revoke → hidden again.
	if err := cols.RemoveShare(ctx, owner.ID, col.ID, guest.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, _, err := bm.ListInCollection(ctx, guest.ID, col.ID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("guest after revoke: got %v, want ErrNotFound", err)
	}
}

func TestAdminOverridePolicyGatesAllCollections(t *testing.T) {
	auth, repos := newTestAuth(t)
	bm := newBookmarkService(repos, nil, auth)
	cols := newCollectionService(repos, nil)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	admin := makeUser(t, repos, "admin", "admin@example.com")
	// Give admin the owner role (which holds collections.manage_all).
	if err := repos.Roles.AssignUserRole(ctx, admin.ID, model.RoleIDOwner, "admin"); err != nil {
		t.Fatalf("assign owner role: %v", err)
	}

	col, err := cols.Create(ctx, owner.ID, CollectionInput{Name: "Private"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Policy off → admin has no access.
	if _, _, err := bm.ListInCollection(ctx, admin.ID, col.ID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("admin with policy off: got %v, want ErrNotFound", err)
	}

	// Turn the policy on → admin gets manage on every collection.
	if err := cols.SetPolicy(ctx, admin.ID, model.PolicyAdminManageAllCollections, true); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	if _, _, err := bm.ListInCollection(ctx, admin.ID, col.ID, repository.BookmarkFilter{Limit: 10}); err != nil {
		t.Fatalf("admin with policy on: %v", err)
	}
}
