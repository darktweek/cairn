package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
)

// Supported extensions and their magic bytes.
var allowedTypes = map[string][]byte{
	".jpg":  {0xFF, 0xD8, 0xFF},
	".jpeg": {0xFF, 0xD8, 0xFF},
	".png":  {0x89, 0x50, 0x4E, 0x47},
	".webp": {0x52, 0x49, 0x46, 0x46},
	".avif": {0x00, 0x00, 0x00}, // ftyp box — checked separately
	".mp4":  {0x00, 0x00, 0x00}, // ftyp box — checked separately
	".webm": {0x1A, 0x45, 0xDF, 0xA3},
}

type WallpaperService interface {
	Upload(ctx context.Context, userID, filename string, data []byte) (*model.Wallpaper, error)
	Delete(ctx context.Context, userID, wallpaperID string) error
	List(ctx context.Context, userID string) ([]*model.Wallpaper, error)
	SetPinned(ctx context.Context, userID, wallpaperID string, pinned bool) error
	UpdateSort(ctx context.Context, userID string, ids []string) error
}

type wallpaperService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func newWallpaperService(repos *repository.Repositories, cfg *config.Config) WallpaperService {
	return &wallpaperService{repos: repos, cfg: cfg}
}

func (s *wallpaperService) Upload(ctx context.Context, userID, originalFilename string, data []byte) (*model.Wallpaper, error) {
	// Check size limit.
	if int64(len(data)) > s.cfg.MaxUploadSize {
		return nil, fmt.Errorf("%w: file too large", ErrInvalidInput)
	}

	// Check user wallpaper quota.
	count, err := s.repos.Wallpapers.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	limit := s.cfg.DefaultWallpaperLimit
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err == nil && user.WallpaperLimit != nil {
		limit = *user.WallpaperLimit
	}

	if count >= limit {
		return nil, ErrWallpaperLimit
	}

	// Validate extension.
	ext := strings.ToLower(filepath.Ext(originalFilename))
	magic, ok := allowedTypes[ext]
	if !ok {
		return nil, ErrUnsupportedFile
	}

	// Validate magic bytes.
	if !validateMagic(data, ext, magic) {
		return nil, fmt.Errorf("%w: file content does not match extension", ErrUnsupportedFile)
	}

	// Generate safe filename.
	id := ulid.Make().String()
	safeFilename := id + ext

	// Ensure user media directory exists.
	userDir := filepath.Join(s.cfg.MediaPath, userID)
	if err := os.MkdirAll(userDir, 0o750); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}

	// Write file.
	dest := filepath.Join(userDir, safeFilename)
	if err := os.WriteFile(dest, data, 0o640); err != nil {
		return nil, fmt.Errorf("write wallpaper: %w", err)
	}

	// Insert in DB.
	w := &model.Wallpaper{
		ID:        id,
		UserID:    userID,
		Filename:  safeFilename,
		CreatedAt: time.Now(),
	}

	if err := s.repos.Wallpapers.Create(ctx, w); err != nil {
		_ = os.Remove(dest)
		return nil, err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &userID,
		Action: "wallpaper_upload",
		Metadata: map[string]any{
			"filename": safeFilename,
		},
		CreatedAt: time.Now(),
	})

	return w, nil
}

func (s *wallpaperService) Delete(ctx context.Context, userID, wallpaperID string) error {
	w, err := s.repos.Wallpapers.GetByID(ctx, wallpaperID, userID)
	if err != nil {
		return ErrNotFound
	}

	// Remove physical file.
	dest := filepath.Join(s.cfg.MediaPath, userID, w.Filename)
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		slog.Error("remove wallpaper file", "path", dest, "err", err)
	}

	if err := s.repos.Wallpapers.Delete(ctx, wallpaperID, userID); err != nil {
		return err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &userID,
		Action: "wallpaper_delete",
		Metadata: map[string]any{
			"filename": w.Filename,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *wallpaperService) List(ctx context.Context, userID string) ([]*model.Wallpaper, error) {
	return s.repos.Wallpapers.ListByUser(ctx, userID)
}

func (s *wallpaperService) SetPinned(ctx context.Context, userID, wallpaperID string, pinned bool) error {
	return s.repos.Wallpapers.SetPinned(ctx, wallpaperID, userID, pinned)
}

func (s *wallpaperService) UpdateSort(ctx context.Context, userID string, ids []string) error {
	return s.repos.Wallpapers.UpdateSort(ctx, userID, ids)
}

// validateMagic checks file header bytes match the expected type.
func validateMagic(data []byte, ext string, magic []byte) bool {
	if len(data) < 12 {
		return false
	}

	switch ext {
	case ".avif", ".mp4":
		// ftyp box: bytes 4-7 must be "ftyp"
		return string(data[4:8]) == "ftyp"
	case ".webp":
		// RIFF....WEBP
		return bytes.HasPrefix(data, magic) && len(data) > 11 && string(data[8:12]) == "WEBP"
	default:
		return bytes.HasPrefix(data, magic)
	}
}
