package repository

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/darktweek/cairn/internal/model"
)

func TestGetOrCreatePersonalIsStable(t *testing.T) {
	repos := newTestRepos(t)
	u := newTestUser(t, repos, "alice", "alice@example.com")
	ctx := context.Background()

	c1, err := repos.Collections.GetOrCreatePersonal(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetOrCreatePersonal: %v", err)
	}
	if !c1.IsPersonal || c1.OwnerID != u.ID {
		t.Fatalf("unexpected personal collection: %+v", c1)
	}

	c2, err := repos.Collections.GetOrCreatePersonal(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetOrCreatePersonal (2nd): %v", err)
	}
	if c2.ID != c1.ID {
		t.Fatalf("personal collection not stable: %s != %s", c1.ID, c2.ID)
	}
}

func TestEffectivePermOwnerVsStranger(t *testing.T) {
	repos := newTestRepos(t)
	owner := newTestUser(t, repos, "owner", "owner@example.com")
	stranger := newTestUser(t, repos, "stranger", "stranger@example.com")
	ctx := context.Background()

	c, err := repos.Collections.GetOrCreatePersonal(ctx, owner.ID)
	if err != nil {
		t.Fatalf("personal: %v", err)
	}

	perm, err := repos.Collections.EffectivePerm(ctx, owner.ID, c.ID)
	if err != nil {
		t.Fatalf("EffectivePerm owner: %v", err)
	}
	if perm != model.PermManage {
		t.Fatalf("owner perm = %q, want manage", perm)
	}

	perm, err = repos.Collections.EffectivePerm(ctx, stranger.ID, c.ID)
	if err != nil {
		t.Fatalf("EffectivePerm stranger: %v", err)
	}
	if perm != "" {
		t.Fatalf("stranger perm = %q, want empty (no access)", perm)
	}
}

func TestListByCollectionReturnsAuthor(t *testing.T) {
	repos := newTestRepos(t)
	u := newTestUser(t, repos, "carol", "carol@example.com")
	ctx := context.Background()

	c, err := repos.Collections.GetOrCreatePersonal(ctx, u.ID)
	if err != nil {
		t.Fatalf("personal: %v", err)
	}

	f := &model.Folder{ID: ulid.Make().String(), CollectionID: c.ID, Name: "News", CreatedAt: time.Now()}
	if err := repos.Folders.Create(ctx, f); err != nil {
		t.Fatalf("folder create: %v", err)
	}

	now := time.Now()
	b := &model.Bookmark{
		ID:           ulid.Make().String(),
		UserID:       u.ID,
		CollectionID: c.ID,
		FolderID:     &f.ID,
		URL:          "https://example.com",
		Title:        "Example",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repos.Bookmarks.Create(ctx, b); err != nil {
		t.Fatalf("bookmark create: %v", err)
	}

	list, total, err := repos.Bookmarks.ListByCollection(ctx, c.ID, BookmarkFilter{Limit: 50})
	if err != nil {
		t.Fatalf("ListByCollection: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("got %d bookmarks (total %d), want 1", len(list), total)
	}
	if list[0].AddedByUsername != "carol" {
		t.Fatalf("AddedByUsername = %q, want carol", list[0].AddedByUsername)
	}
	if list[0].FolderID == nil || *list[0].FolderID != f.ID {
		t.Fatalf("folder id not preserved: %+v", list[0].FolderID)
	}

	// Folder filter narrows correctly.
	filtered, _, err := repos.Bookmarks.ListByCollection(ctx, c.ID, BookmarkFilter{Limit: 50, FolderID: &f.ID})
	if err != nil {
		t.Fatalf("ListByCollection filtered: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("folder filter got %d, want 1", len(filtered))
	}
}
