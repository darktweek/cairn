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
	SetUploadSizeLimit(ctx context.Context, adminID, userID string, limit *int64) error
	SetStorageQuota(ctx context.Context, adminID, userID string, quota *int64) error
	ListPendingRegistrations(ctx context.Context) ([]*model.PendingRegistration, error)
	RevokePendingRegistration(ctx context.Context, adminID, id string) error
	GetAuditLog(ctx context.Context, offset, limit int, filter repository.AuditFilter) ([]*model.AuditEntry, int, error)
	ResolveUsernames(ctx context.Context, ids []string) (map[string]string, error)
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

	if err := s.repos.Users.HardDelete(ctx, userID); err != nil {
		return err
	}

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

func (s *adminService) SetUploadSizeLimit(ctx context.Context, adminID, userID string, limit *int64) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.UploadSizeLimit = limit
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *adminService) SetStorageQuota(ctx context.Context, adminID, userID string, quota *int64) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	user.StorageQuota = quota
	user.UpdatedAt = time.Now()

	return s.repos.Users.Update(ctx, user)
}

func (s *adminService) ListPendingRegistrations(ctx context.Context) ([]*model.PendingRegistration, error) {
	return s.repos.PendingRegistrations.List(ctx)
}

func (s *adminService) RevokePendingRegistration(ctx context.Context, adminID, id string) error {
	if err := s.repos.PendingRegistrations.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &adminID,
		Action: "registration_revoked",
		Metadata: map[string]any{
			"pending_registration_id": id,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *adminService) GetAuditLog(ctx context.Context, offset, limit int, filter repository.AuditFilter) ([]*model.AuditEntry, int, error) {
	return s.repos.Audit.List(ctx, offset, limit, filter)
}

func (s *adminService) ResolveUsernames(ctx context.Context, ids []string) (map[string]string, error) {
	return s.repos.Users.UsernamesByIDs(ctx, ids)
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
	activeCount, bookmarks, wallpapers, sessions := 0, 0, 0, 0
	for _, u := range allUsers {
		if u.IsActive {
			activeCount++
		}
		if n, err := s.repos.Bookmarks.CountByUser(ctx, u.ID); err == nil {
			bookmarks += n
		}
		if n, err := s.repos.Wallpapers.CountByUser(ctx, u.ID); err == nil {
			wallpapers += n
		}
		if ss, err := s.repos.Sessions.ListByUserID(ctx, u.ID); err == nil {
			sessions += len(ss)
		}
	}

	// DB file size.
	var dbSize int64
	if info, err := os.Stat(s.cfg.DBPath); err == nil {
		dbSize = info.Size()
	}

	pendingInv := 0
	if invs, err := s.repos.Invitations.List(ctx); err == nil {
		for _, inv := range invs {
			if inv.IsPending() {
				pendingInv++
			}
		}
	}

	pendingReg := 0
	if regs, err := s.repos.PendingRegistrations.List(ctx); err == nil {
		for _, r := range regs {
			if !r.IsExpired() {
				pendingReg++
			}
		}
	}

	auditTotal := 0
	if _, n, err := s.repos.Audit.List(ctx, 0, 1, repository.AuditFilter{}); err == nil {
		auditTotal = n
	}

	stats := &model.AdminStats{
		TotalUsers:           total,
		ActiveUsers:          activeCount,
		TotalBookmarks:       bookmarks,
		TotalWallpapers:      wallpapers,
		DBSizeBytes:          dbSize,
		MediaBytes:           dirSize(s.cfg.MediaPath),
		ActiveSessions:       sessions,
		PendingInvitations:   pendingInv,
		PendingRegistrations: pendingReg,
		AuditEntries:         auditTotal,
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
