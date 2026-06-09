package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

type AdminService interface {
	ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int, error)
	GetUser(ctx context.Context, userID string) (*model.User, error)
	SuspendUser(ctx context.Context, adminID, userID string) error
	ActivateUser(ctx context.Context, adminID, userID string) error
	DeleteUser(ctx context.Context, adminID, userID string) error
	SetWallpaperLimit(ctx context.Context, adminID, userID string, limit *int) error
	GetAuditLog(ctx context.Context, offset, limit int, filter repository.AuditFilter) ([]*model.AuditEntry, int, error)
	GetStats(ctx context.Context) (*model.AdminStats, error)
}

type adminService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func newAdminService(repos *repository.Repositories, cfg *config.Config) AdminService {
	return &adminService{repos: repos, cfg: cfg}
}

func (s *adminService) ListUsers(ctx context.Context, offset, limit int) ([]*model.User, int, error) {
	return s.repos.Users.List(ctx, offset, limit)
}

func (s *adminService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	u, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *adminService) SuspendUser(ctx context.Context, adminID, userID string) error {
	if adminID == userID {
		return fmt.Errorf("%w: cannot suspend yourself", ErrForbidden)
	}

	if err := s.guardLastAdmin(ctx, userID); err != nil {
		return err
	}

	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	if err := s.repos.Users.Update(ctx, user); err != nil {
		return err
	}

	_ = s.repos.Sessions.DeleteByUserID(ctx, userID)

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &adminID,
		Action: "user_suspended",
		Metadata: map[string]any{
			"target_user_id": userID,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *adminService) ActivateUser(ctx context.Context, adminID, userID string) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.IsActive = true
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *adminService) DeleteUser(ctx context.Context, adminID, userID string) error {
	if adminID == userID {
		return fmt.Errorf("%w: cannot delete yourself", ErrForbidden)
	}

	if err := s.guardLastAdmin(ctx, userID); err != nil {
		return err
	}

	// Remove all physical media files.
	userDir := filepath.Join(s.cfg.MediaPath, userID)
	_ = os.RemoveAll(userDir)

	if err := s.repos.Users.SoftDelete(ctx, userID); err != nil {
		return err
	}

	_ = s.repos.Sessions.DeleteByUserID(ctx, userID)

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &adminID,
		Action: "user_deleted",
		Metadata: map[string]any{
			"target_user_id": userID,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *adminService) SetWallpaperLimit(ctx context.Context, adminID, userID string, limit *int) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.WallpaperLimit = limit
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *adminService) GetAuditLog(ctx context.Context, offset, limit int, filter repository.AuditFilter) ([]*model.AuditEntry, int, error) {
	return s.repos.Audit.List(ctx, offset, limit, filter)
}

func (s *adminService) GetStats(ctx context.Context) (*model.AdminStats, error) {
	total, err := s.repos.Users.Count(ctx)
	if err != nil {
		return nil, err
	}

	allUsers, _, err := s.repos.Users.List(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}
	activeCount := 0
	for _, u := range allUsers {
		if u.IsActive {
			activeCount++
		}
	}

	// DB file size.
	var dbSize int64
	if info, err := os.Stat(s.cfg.DBPath); err == nil {
		dbSize = info.Size()
	}

	stats := &model.AdminStats{
		TotalUsers:  total,
		ActiveUsers: activeCount,
		DBSizeBytes: dbSize,
	}

	return stats, nil
}

// guardLastAdmin returns ErrForbidden if the target user is the only active admin.
func (s *adminService) guardLastAdmin(ctx context.Context, targetUserID string) error {
	target, err := s.repos.Users.GetByID(ctx, targetUserID)
	if err != nil {
		return ErrNotFound
	}

	if target.Role != "admin" {
		return nil
	}

	allUsers, _, err := s.repos.Users.List(ctx, 0, 10000)
	if err != nil {
		return err
	}

	adminCount := 0
	for _, u := range allUsers {
		if u.Role == "admin" && u.IsActive {
			adminCount++
		}
	}

	if adminCount <= 1 {
		return fmt.Errorf("%w: cannot remove the last admin", ErrForbidden)
	}

	return nil
}
