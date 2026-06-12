package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/darktweek/cairn/internal/db"
	"github.com/darktweek/cairn/internal/model"
)

func newTestRepos(t *testing.T) *Repositories {
	t.Helper()
	database, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(database)
}

func newTestUser(t *testing.T, repos *Repositories, username, email string) *model.User {
	t.Helper()
	now := time.Now()
	u := &model.User{
		ID:        ulid.Make().String(),
		Username:  username,
		Email:     email,
		Password:     "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		Role:         "user",
		IsActive:     true,
		SearchEngine: "duckduckgo",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Users.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestGetByEmailCaseInsensitive(t *testing.T) {
	repos := newTestRepos(t)
	newTestUser(t, repos, "alice", "alice@example.com")

	for _, email := range []string{"alice@example.com", "Alice@Example.com", "ALICE@EXAMPLE.COM"} {
		u, err := repos.Users.GetByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("GetByEmail(%q): %v", email, err)
		}
		if u.Username != "alice" {
			t.Fatalf("GetByEmail(%q): got %q", email, u.Username)
		}
	}
}

func TestGetByUsernameCaseInsensitive(t *testing.T) {
	repos := newTestRepos(t)
	newTestUser(t, repos, "Bob", "bob@example.com")

	for _, name := range []string{"Bob", "bob", "BOB"} {
		u, err := repos.Users.GetByUsername(context.Background(), name)
		if err != nil {
			t.Fatalf("GetByUsername(%q): %v", name, err)
		}
		if u.Email != "bob@example.com" {
			t.Fatalf("GetByUsername(%q): got %q", name, u.Email)
		}
	}
}

func TestGetByEmailUnknown(t *testing.T) {
	repos := newTestRepos(t)
	if _, err := repos.Users.GetByEmail(context.Background(), "ghost@example.com"); err == nil {
		t.Fatal("unknown email must return an error")
	}
}

func TestSoftDeletedUserIsInvisible(t *testing.T) {
	repos := newTestRepos(t)
	u := newTestUser(t, repos, "carol", "carol@example.com")

	if err := repos.Users.SoftDelete(context.Background(), u.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if _, err := repos.Users.GetByEmail(context.Background(), "carol@example.com"); err == nil {
		t.Fatal("soft-deleted user must not be returned by GetByEmail")
	}
}
