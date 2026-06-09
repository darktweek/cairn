package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

type UserService interface {
	Register(ctx context.Context, username, email, password, ip, userAgent string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID, username, email string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	UpdateSearchEngine(ctx context.Context, userID, engine string, customURL *string) error
	GetAuditLog(ctx context.Context, userID string, offset, limit int) ([]*model.AuditEntry, int, error)
}

type userService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func newUserService(repos *repository.Repositories, cfg *config.Config) UserService {
	return &userService{repos: repos, cfg: cfg}
}

func (s *userService) Register(ctx context.Context, username, email, password, ip, userAgent string) (*model.User, error) {
	if !usernameRe.MatchString(username) {
		return nil, fmt.Errorf("%w: username must be 3-32 alphanumeric characters", ErrInvalidInput)
	}
	if !isValidEmail(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}
	if len(password) < 12 {
		return nil, fmt.Errorf("%w: password must be at least 12 characters", ErrInvalidInput)
	}

	isFirst, err := s.repos.Users.IsFirstUser(ctx)
	if err != nil {
		return nil, err
	}

	role := "user"
	if isFirst {
		role = "admin"
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &model.User{
		ID:           ulid.Make().String(),
		Username:     username,
		Email:        strings.ToLower(email),
		Password:     hashed,
		Role:         role,
		IsActive:     true,
		SearchEngine: "duckduckgo",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repos.Users.Create(ctx, user); err != nil {
		// Anti-enumeration: never reveal which field collided. Log the real
		// reason to the audit trail (visible to admins) but return a generic
		// error to the client.
		if existing, e := s.repos.Users.GetByEmail(ctx, user.Email); e == nil && existing != nil {
			email := user.Email
			_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
				ID:        ulid.Make().String(),
				Action:    "register_blocked_duplicate_email",
				IP:        ip,
				UserAgent: userAgent,
				Metadata:  map[string]any{"email": email},
				CreatedAt: now,
			})
		}
		return nil, fmt.Errorf("%w: impossible de créer le compte avec ces informations", ErrConflict)
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &user.ID,
		Action:    "user_created",
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: now,
	})

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	u, err := s.repos.Users.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID, username, email string) error {
	if !usernameRe.MatchString(username) {
		return fmt.Errorf("%w: invalid username", ErrInvalidInput)
	}
	if !isValidEmail(email) {
		return fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}

	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.Username = username
	user.Email = strings.ToLower(email)
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *userService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) < 12 {
		return fmt.Errorf("%w: password too short", ErrInvalidInput)
	}

	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	if !verifyPassword(currentPassword, user.Password) {
		return ErrUnauthorized
	}

	hashed, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashed
	user.UpdatedAt = time.Now()

	if err := s.repos.Users.Update(ctx, user); err != nil {
		return err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "password_change",
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *userService) UpdateSearchEngine(ctx context.Context, userID, engine string, customURL *string) error {
	valid := map[string]bool{
		"duckduckgo": true, "google": true, "brave": true,
		"bing": true, "kagi": true, "custom": true,
	}
	if !valid[engine] {
		return fmt.Errorf("%w: unknown search engine", ErrInvalidInput)
	}
	if engine == "custom" {
		if customURL == nil || *customURL == "" {
			return fmt.Errorf("%w: custom URL required", ErrInvalidInput)
		}
		if !strings.HasSuffix(*customURL, "=") {
			return fmt.Errorf("%w: custom URL must end with '='", ErrInvalidInput)
		}
	}

	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.SearchEngine = engine
	user.SearchEngineURL = customURL
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *userService) GetAuditLog(ctx context.Context, userID string, offset, limit int) ([]*model.AuditEntry, int, error) {
	return s.repos.Audit.ListByUser(ctx, userID, offset, limit)
}

// isValidEmail is a simple RFC 5322 sanity check — no external dep.
func isValidEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 1 || at >= len(email)-1 {
		return false
	}
	domain := email[at+1:]
	return strings.Contains(domain, ".") && len(domain) > 2
}
