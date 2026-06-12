package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/db"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
)

func newTestAuth(t *testing.T) (AuthService, *repository.Repositories) {
	t.Helper()
	database, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repos := repository.New(database)
	cfg := &config.Config{}
	settings := newSettingsService(repos, cfg)
	email := newEmailService(cfg, settings)
	return newAuthService(repos, cfg, settings, email), repos
}

func createTestUser(t *testing.T, repos *repository.Repositories, username, email, password string) {
	t.Helper()
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	now := time.Now()
	u := &model.User{
		ID:        ulid.Make().String(),
		Username:  username,
		Email:     email,
		Password:     hash,
		Role:         "user",
		IsActive:     true,
		SearchEngine: "duckduckgo",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Users.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func TestLoginWithEmailAndUsername(t *testing.T) {
	auth, repos := newTestAuth(t)
	createTestUser(t, repos, "dave", "dave@example.com", "s3cret-pass")

	// All of these must reach the same account.
	for _, id := range []string{"dave@example.com", "DAVE@EXAMPLE.COM", "dave", " Dave@example.com "} {
		sess, token, err := auth.Login(context.Background(), id, "s3cret-pass", "", "127.0.0.1", "test")
		if err != nil {
			t.Fatalf("Login(%q): %v", id, err)
		}
		if sess == nil || token == "" {
			t.Fatalf("Login(%q): missing session or token", id)
		}
	}
}

func TestLoginWrongPassword(t *testing.T) {
	auth, repos := newTestAuth(t)
	createTestUser(t, repos, "erin", "erin@example.com", "right-pass")

	_, _, err := auth.Login(context.Background(), "erin@example.com", "wrong-pass", "", "127.0.0.1", "test")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
}

func TestLoginUnknownIdentifier(t *testing.T) {
	auth, _ := newTestAuth(t)
	_, _, err := auth.Login(context.Background(), "ghost", "x", "", "127.0.0.1", "test")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
}

func TestLoginPerAccountRateLimit(t *testing.T) {
	auth, repos := newTestAuth(t)
	createTestUser(t, repos, "frank", "frank@example.com", "right-pass")
	createTestUser(t, repos, "grace", "grace@example.com", "right-pass")

	// Burn frank's bucket with bad attempts.
	for i := 0; i < 10; i++ {
		auth.Login(context.Background(), "frank@example.com", "bad", "", "127.0.0.1", "test")
	}
	_, _, err := auth.Login(context.Background(), "frank@example.com", "right-pass", "", "127.0.0.1", "test")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("frank should be rate limited, got %v", err)
	}

	// Same IP, different account: must not be affected.
	if _, _, err := auth.Login(context.Background(), "grace@example.com", "right-pass", "", "127.0.0.1", "test"); err != nil {
		t.Fatalf("grace must not share frank's bucket: %v", err)
	}
}

func TestLoginInactiveUser(t *testing.T) {
	auth, repos := newTestAuth(t)
	createTestUser(t, repos, "henry", "henry@example.com", "right-pass")

	u, err := repos.Users.GetByEmail(context.Background(), "henry@example.com")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	u.IsActive = false
	if err := repos.Users.Update(context.Background(), u); err != nil {
		t.Fatalf("update: %v", err)
	}

	_, _, err = auth.Login(context.Background(), "henry@example.com", "right-pass", "", "127.0.0.1", "test")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("inactive user must get ErrUnauthorized, got %v", err)
	}
}
