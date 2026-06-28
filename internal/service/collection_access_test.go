package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
)

func makeUser(t *testing.T, repos *repository.Repositories, username, email string) *model.User {
	t.Helper()
	now := time.Now()
	u := &model.User{
		ID:           ulid.Make().String(),
		Username:     username,
		Email:        email,
		Password:     "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		Role:         "user",
		IsActive:     true,
		SearchEngine: "duckduckgo",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repos.Users.Create(context.Background(), u); err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return u
}

// A stranger must not be able to read or write another user's collection.
func TestBookmarkServiceEnforcesCollectionAccess(t *testing.T) {
	auth, repos := newTestAuth(t)
	svc := newBookmarkService(repos, &config.Config{}, auth)
	ctx := context.Background()

	owner := makeUser(t, repos, "owner", "owner@example.com")
	stranger := makeUser(t, repos, "stranger", "stranger@example.com")

	// Owner creates a bookmark in their (lazily-created) personal collection.
	bm, err := svc.Create(ctx, owner.ID, BookmarkInput{URL: "https://example.com", Title: "Ex"})
	if err != nil {
		t.Fatalf("owner create: %v", err)
	}
	if bm.CollectionID == "" {
		t.Fatal("created bookmark has no collection")
	}

	// A stranger has no access at all, so every operation is hidden as ErrNotFound.
	_, err = svc.Create(ctx, stranger.ID, BookmarkInput{
		URL: "https://evil.com", Title: "X", CollectionID: bm.CollectionID,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("stranger create into owner collection: got %v, want ErrNotFound", err)
	}

	if err := svc.Delete(ctx, stranger.ID, bm.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stranger delete: got %v, want ErrNotFound", err)
	}

	if _, _, err := svc.ListInCollection(ctx, stranger.ID, bm.CollectionID, repository.BookmarkFilter{Limit: 10}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stranger list: got %v, want ErrNotFound", err)
	}

	// Owner can still list their own collection and see the bookmark.
	list, total, err := svc.ListInCollection(ctx, owner.ID, bm.CollectionID, repository.BookmarkFilter{Limit: 10})
	if err != nil {
		t.Fatalf("owner list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("owner list got %d (total %d), want 1", len(list), total)
	}
}
