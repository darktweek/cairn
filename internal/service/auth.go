package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/ratelimit"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/argon2"
)

const (
	sessionLifetime    = 30 * 24 * time.Hour
	resetTokenLifetime = 1 * time.Hour
	argon2Time         = 1
	argon2Memory       = 64 * 1024 // 64 MB
	argon2Threads      = 4
	argon2KeyLen       = 32
	argon2SaltLen      = 16
)

type AuthService interface {
	Login(ctx context.Context, email, password, totpCode, ip, userAgent string) (*model.Session, string, error)
	CreateSessionForUser(ctx context.Context, userID, ip, userAgent string) (*model.Session, string, error)
	Logout(ctx context.Context, sessionID, userID string) error
	LogoutForUser(ctx context.Context, sessionID, userID string) error
	LogoutAll(ctx context.Context, userID string) error
	ListSessions(ctx context.Context, userID string) ([]*model.Session, error)
	ValidateSession(ctx context.Context, token string) (*model.User, *model.Session, error)
	CreateBookmarkletSession(ctx context.Context, userID, ip, userAgent string) (string, error)
	BeginTOTP(ctx context.Context, userID string) (secret, qrCodeURL, qrImage string, err error)
	ConfirmTOTP(ctx context.Context, userID, code string) error
	ValidateTOTP(ctx context.Context, userID, code string) (bool, error)
	DisableTOTP(ctx context.Context, userID, password string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	// Registration flow (email-verified, TOTP-mandatory)
	RequestRegistration(ctx context.Context, username, email string) error
	ValidateSetupToken(ctx context.Context, token string) (username, email, totpSecret, totpURI, qrImage string, err error)
	CompleteSetup(ctx context.Context, token, password, totpCode string, ip, userAgent string) (*model.Session, string, error)
	// Invite setup flow (for invite tokens)
	PrepareInviteSetup(ctx context.Context, inviteToken, username string) (email, totpSecret, totpURI, qrImage string, err error)
	CompleteInviteSetup(ctx context.Context, inviteToken, password, totpCode string, ip, userAgent string) (*model.Session, string, error)
	// Account deletion
	DeleteAccount(ctx context.Context, userID, password string) error
}

type resetEntry struct {
	userID    string
	expiresAt time.Time
}

type authService struct {
	repos       *repository.Repositories
	cfg         *config.Config
	settings    SettingsService
	email       EmailService
	resetTokens sync.Map // tokenHash → resetEntry
	// Per-account login throttle — independent from the per-IP middleware,
	// which collapses to a single shared bucket behind Docker NAT.
	loginLimiter *ratelimit.Limiter
}

func newAuthService(repos *repository.Repositories, cfg *config.Config, settings SettingsService, email EmailService) AuthService {
	return &authService{
		repos: repos, cfg: cfg, settings: settings, email: email,
		loginLimiter: ratelimit.New(ratelimit.Config{Max: 10, Window: 5 * time.Minute}),
	}
}

// Login authenticates a user. totpCode is optional — if TOTP is enabled and empty, returns ErrTOTPRequired.
func (s *authService) Login(ctx context.Context, email, password, totpCode, ip, userAgent string) (*model.Session, string, error) {
	// Mobile keyboards autocapitalize and autofill can pad — never let
	// formatting cost someone their login.
	email = strings.ToLower(strings.TrimSpace(email))

	// Throttle per submitted identifier: brute-forcing one account stays
	// impossible even when every client shares the proxy/NAT IP, and one
	// noisy device can't lock the whole network out.
	if !s.loginLimiter.Allow(email) {
		return nil, "", ErrRateLimited
	}

	user, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		// The identifier field accepts a username too.
		user, err = s.repos.Users.GetByUsername(ctx, email)
	}
	if err != nil {
		// avoid enumeration: same error for not found or wrong password
		return nil, "", ErrUnauthorized
	}

	if !user.IsActive {
		return nil, "", ErrUnauthorized
	}

	if !verifyPassword(password, user.Password) {
		_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
			ID:        ulid.Make().String(),
			UserID:    &user.ID,
			Action:    "login_failed",
			IP:        ip,
			UserAgent: userAgent,
			CreatedAt: time.Now(),
		})
		return nil, "", ErrUnauthorized
	}

	totp, totpErr := s.repos.TOTP.GetByUserID(ctx, user.ID)
	if totpErr == nil && totp.IsVerified {
		if totpCode == "" {
			return nil, "", ErrTOTPRequired
		}
		secret, err := s.decryptAESGCM(totp.Secret)
		if err != nil || !validateTOTPCode(string(secret), totpCode, time.Now()) {
			return nil, "", ErrInvalidTOTP
		}
	}

	sess, token, err := s.createSession(ctx, user.ID, ip, userAgent, false)
	if err != nil {
		return nil, "", err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &user.ID,
		Action:    "login",
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	})

	return sess, token, nil
}

// CreateSessionForUser issues a session for an already-authenticated user
// (used by the SSO callback after the provider verified the identity).
func (s *authService) CreateSessionForUser(ctx context.Context, userID, ip, userAgent string) (*model.Session, string, error) {
	sess, token, err := s.createSession(ctx, userID, ip, userAgent, false)
	if err != nil {
		return nil, "", err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "login_sso",
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	})
	return sess, token, nil
}

func (s *authService) Logout(ctx context.Context, sessionID, userID string) error {
	if err := s.repos.Sessions.DeleteByID(ctx, sessionID); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "logout",
		CreatedAt: time.Now(),
	})
	return nil
}

// LogoutForUser revokes a session only if it belongs to the given user — prevents IDOR.
func (svc *authService) LogoutForUser(ctx context.Context, sessionID, userID string) error {
	sessions, err := svc.repos.Sessions.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		if sess.ID == sessionID {
			if err := svc.repos.Sessions.DeleteByID(ctx, sessionID); err != nil {
				return err
			}
			_ = svc.repos.Audit.Log(ctx, &model.AuditEntry{
				ID:        ulid.Make().String(),
				UserID:    &userID,
				Action:    "session_revoked",
				CreatedAt: time.Now(),
			})
			return nil
		}
	}
	return ErrNotFound
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	if err := s.repos.Sessions.DeleteByUserID(ctx, userID); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "logout_all",
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *authService) ListSessions(ctx context.Context, userID string) ([]*model.Session, error) {
	return s.repos.Sessions.ListByUserID(ctx, userID)
}

func (s *authService) ValidateSession(ctx context.Context, token string) (*model.User, *model.Session, error) {
	hash := hashToken(token)

	sess, err := s.repos.Sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}

	user, err := s.repos.Users.GetByID(ctx, sess.UserID)
	if err != nil || !user.IsActive {
		return nil, nil, ErrUnauthorized
	}

	// Load the instance permission set so downstream RequirePermission checks
	// and the frontend can reason about what this user may do.
	if perms, err := s.repos.Roles.PermissionsForUser(ctx, user.ID); err == nil {
		user.Permissions = perms
	}
	if roles, err := s.repos.Roles.RolesForUser(ctx, user.ID); err == nil {
		user.Roles = roles
	}
	if user.RoleName == "" {
		user.RoleName = "user"
	}

	return user, sess, nil
}

func (s *authService) BeginTOTP(ctx context.Context, userID string) (string, string, string, error) {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return "", "", "", ErrNotFound
	}

	secret, err := generateTOTPSecret()
	if err != nil {
		return "", "", "", fmt.Errorf("generate totp secret: %w", err)
	}

	encrypted, err := s.encryptAESGCM([]byte(secret))
	if err != nil {
		return "", "", "", fmt.Errorf("encrypt totp secret: %w", err)
	}

	if err := s.repos.TOTP.Create(ctx, userID, encrypted); err != nil {
		return "", "", "", err
	}

	qrURL := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(s.settings.TOTPIssuer(ctx).Value),
		url.PathEscape(user.Email),
		secret,
		url.QueryEscape(s.settings.TOTPIssuer(ctx).Value),
	)

	return secret, qrURL, totpQRImage(qrURL), nil
}

func (s *authService) ConfirmTOTP(ctx context.Context, userID, code string) error {
	totp, err := s.repos.TOTP.GetByUserID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}

	secret, err := s.decryptAESGCM(totp.Secret)
	if err != nil {
		return fmt.Errorf("decrypt totp: %w", err)
	}

	if !validateTOTPCode(string(secret), code, time.Now()) {
		return ErrInvalidTOTP
	}

	if err := s.repos.TOTP.Verify(ctx, userID); err != nil {
		return err
	}
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "totp_enabled",
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *authService) ValidateTOTP(ctx context.Context, userID, code string) (bool, error) {
	totp, err := s.repos.TOTP.GetByUserID(ctx, userID)
	if err != nil {
		return false, ErrNotFound
	}
	if !totp.IsVerified {
		return false, ErrInvalidTOTP
	}

	secret, err := s.decryptAESGCM(totp.Secret)
	if err != nil {
		return false, fmt.Errorf("decrypt totp: %w", err)
	}

	return validateTOTPCode(string(secret), code, time.Now()), nil
}

func (s *authService) DisableTOTP(ctx context.Context, userID, password string) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if !verifyPassword(password, user.Password) {
		return ErrUnauthorized
	}

	if err := s.repos.TOTP.Delete(ctx, userID); err != nil {
		return err
	}

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &userID,
		Action:    "totp_disabled",
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// Always return nil to avoid email enumeration — caller always responds 202
	user, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	token, err := generateRawToken()
	if err != nil {
		slog.Error("forgot password token generation", "err", err)
		return nil
	}

	hash := hashToken(token)
	s.resetTokens.Store(hash, resetEntry{
		userID:    user.ID,
		expiresAt: time.Now().Add(resetTokenLifetime),
	})

	// fire-and-forget — we log "login_failed" as the closest available action.
	// A dedicated "password_reset_requested" action would need a schema migration.

	slog.Info("password reset token generated", "userID", user.ID)
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &user.ID,
		Action:    "password_reset_requested",
		CreatedAt: time.Now(),
	})
	return nil
}

func (s *authService) CreateBookmarkletSession(ctx context.Context, userID, ip, userAgent string) (string, error) {
	_, token, err := s.createSession(ctx, userID, ip, userAgent, true)
	return token, err
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 12 {
		return fmt.Errorf("%w: password too short", ErrInvalidInput)
	}

	hash := hashToken(token)
	v, ok := s.resetTokens.Load(hash)
	if !ok {
		return ErrUnauthorized
	}

	entry := v.(resetEntry)
	if time.Now().After(entry.expiresAt) {
		s.resetTokens.Delete(hash)
		return ErrUnauthorized
	}

	user, err := s.repos.Users.GetByID(ctx, entry.userID)
	if err != nil {
		return ErrNotFound
	}

	hashed, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.Password = hashed
	user.UpdatedAt = time.Now()
	if err := s.repos.Users.Update(ctx, user); err != nil {
		return err
	}

	s.resetTokens.Delete(hash)
	_ = s.repos.Sessions.DeleteByUserID(ctx, user.ID)
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:        ulid.Make().String(),
		UserID:    &user.ID,
		Action:    "password_reset",
		CreatedAt: time.Now(),
	})
	return nil
}

// createSession generates a cryptographically random token, stores its SHA-256 hash.
func (s *authService) createSession(ctx context.Context, userID, ip, userAgent string, isBookmarklet bool) (*model.Session, string, error) {
	token, err := generateRawToken()
	if err != nil {
		return nil, "", err
	}

	lifetime := sessionLifetime
	if days := s.cfg.SessionLifetimeDays; days > 0 {
		lifetime = time.Duration(days) * 24 * time.Hour
	}
	if isBookmarklet {
		lifetime = time.Duration(s.settings.BookmarkletDays(ctx).Value) * 24 * time.Hour
	}

	sess := &model.Session{
		ID:            ulid.Make().String(),
		UserID:        userID,
		TokenHash:     hashToken(token),
		UserAgent:     userAgent,
		IP:            ip,
		ExpiresAt:     time.Now().Add(lifetime),
		CreatedAt:     time.Now(),
		IsBookmarklet: isBookmarklet,
	}

	if err := s.repos.Sessions.Create(ctx, sess); err != nil {
		return nil, "", err
	}

	return sess, token, nil
}

// hashToken returns the hex-encoded SHA-256 of a token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

// generateRawToken returns 32 cryptographically random bytes encoded as base64url.
func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashPassword hashes a plaintext password with argon2id.
func hashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// verifyPassword checks a plaintext password against a stored argon2id hash.
func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}

	var memory, timeCost, threads uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computed := argon2.IDKey([]byte(password), salt, timeCost, memory, uint8(threads), uint32(len(hash)))
	return hmac.Equal(hash, computed)
}

// generateTOTPSecret returns a base32-encoded 20-byte random secret.
func generateTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// validateTOTPCode verifies a 6-digit code against the secret, checking current window ±1.
func validateTOTPCode(secret, code string, t time.Time) bool {
	counter := uint64(t.Unix()) / 30
	for _, c := range []uint64{counter - 1, counter, counter + 1} {
		// Constant-time comparison to prevent timing attacks.
		if subtle.ConstantTimeCompare([]byte(totpCode(secret, c)), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func totpCode(secret string, counter uint64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	code := (uint32(h[offset])&0x7f)<<24 |
		uint32(h[offset+1])<<16 |
		uint32(h[offset+2])<<8 |
		uint32(h[offset+3])

	return fmt.Sprintf("%06d", new(big.Int).SetUint64(uint64(code)).Mod(
		new(big.Int).SetUint64(uint64(code)),
		big.NewInt(1_000_000),
	).Uint64())
}

// encryptAESGCM encrypts plaintext using AES-256-GCM, key derived from session secret.
func (s *authService) encryptAESGCM(plaintext []byte) (string, error) {
	return encryptValue(deriveKey(s.cfg.SessionSecret), string(plaintext))
}

// decryptAESGCM decrypts a base64-encoded AES-GCM ciphertext.
func (s *authService) decryptAESGCM(encoded string) ([]byte, error) {
	pt, err := decryptValue(deriveKey(s.cfg.SessionSecret), encoded)
	return []byte(pt), err
}

const pendingRegLifetime = 24 * time.Hour

// RequestRegistration validates username/email, creates a pending_registration, and sends the setup email.
func (s *authService) RequestRegistration(ctx context.Context, username, email string) error {
	if len(username) < 2 || len(username) > 32 {
		return fmt.Errorf("%w: username must be 2–32 characters", ErrInvalidInput)
	}
	if len(email) < 3 {
		return fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}

	// Check uniqueness (don't reveal which one to avoid enumeration — return generic error).
	if _, err := s.repos.Users.GetByUsername(ctx, username); err == nil {
		return fmt.Errorf("%w: username or email already taken", ErrConflict)
	}
	if _, err := s.repos.Users.GetByEmail(ctx, email); err == nil {
		return fmt.Errorf("%w: username or email already taken", ErrConflict)
	}

	totpSecret, err := generateTOTPSecret()
	if err != nil {
		return fmt.Errorf("generate totp secret: %w", err)
	}
	encryptedTOTP, err := s.encryptAESGCM([]byte(totpSecret))
	if err != nil {
		return fmt.Errorf("encrypt totp: %w", err)
	}

	token, err := generateRawToken()
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	pr := &model.PendingRegistration{
		ID:         ulid.Make().String(),
		Username:   username,
		Email:      email,
		TokenHash:  hashToken(token),
		TOTPSecret: encryptedTOTP,
		ExpiresAt:  time.Now().Add(pendingRegLifetime),
		CreatedAt:  time.Now(),
	}
	if err := s.repos.PendingRegistrations.Create(ctx, pr); err != nil {
		return fmt.Errorf("create pending registration: %w", err)
	}

	setupURL := s.cfg.BaseURL + "/setup?token=" + token

	// For the very first account, SMTP is not required: log the setup URL prominently
	// so the admin can complete setup directly from the container logs.
	userCount, _ := s.repos.Users.Count(ctx)
	if userCount == 0 {
		slog.Warn("┌─────────────────────────────────────────────────────────────┐")
		slog.Warn("│  FIRST ACCOUNT — complete your setup at the URL below       │")
		slog.Warn("│  No SMTP required for this step.                            │")
		slog.Warn("└─────────────────────────────────────────────────────────────┘",
			"setup_url", setupURL,
			"username", username,
			"expires_at", pr.ExpiresAt.Format("2006-01-02 15:04:05"),
		)
	}

	slog.Info("registration requested", "username", username, "email", email, "setup_url", setupURL)
	_ = s.email.SendAccountSetup(ctx, email, username, setupURL, pr.ExpiresAt)

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		Action: "registration_requested",
		Metadata: map[string]any{
			"username": username,
			"email":    email,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

// ValidateSetupToken validates a pending_registration token and returns the setup data.
func (s *authService) ValidateSetupToken(ctx context.Context, token string) (username, email, totpSecret, totpURI, qrImage string, err error) {
	pr, err := s.repos.PendingRegistrations.GetByTokenHash(ctx, hashToken(token))
	if err != nil {
		return "", "", "", "", "", ErrUnauthorized
	}
	if pr.IsExpired() || pr.IsCompleted() {
		return "", "", "", "", "", ErrUnauthorized
	}

	plain, err := s.decryptAESGCM(pr.TOTPSecret)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("decrypt totp: %w", err)
	}
	issuer := s.settings.TOTPIssuer(ctx).Value
	uri := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(issuer), url.PathEscape(pr.Email),
		string(plain), url.QueryEscape(issuer),
	)
	return pr.Username, pr.Email, string(plain), uri, totpQRImage(uri), nil
}

// CompleteSetup finalises a pending_registration: validates TOTP code, creates the user, and opens a session.
func (s *authService) CompleteSetup(ctx context.Context, token, password, totpCode string, ip, userAgent string) (*model.Session, string, error) {
	if len(password) < 12 {
		return nil, "", fmt.Errorf("%w: password must be at least 12 characters", ErrInvalidInput)
	}

	pr, err := s.repos.PendingRegistrations.GetByTokenHash(ctx, hashToken(token))
	if err != nil || pr.IsExpired() || pr.IsCompleted() {
		return nil, "", ErrUnauthorized
	}

	plain, err := s.decryptAESGCM(pr.TOTPSecret)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt totp: %w", err)
	}
	if !validateTOTPCode(string(plain), totpCode, time.Now()) {
		return nil, "", ErrInvalidTOTP
	}

	userSvc := newUserService(s.repos, s.cfg)
	user, err := userSvc.Register(ctx, pr.Username, pr.Email, password, ip, userAgent)
	if err != nil {
		return nil, "", err
	}

	// Store TOTP as verified immediately.
	if err := s.repos.TOTP.Create(ctx, user.ID, pr.TOTPSecret); err != nil {
		return nil, "", fmt.Errorf("create totp: %w", err)
	}
	if err := s.repos.TOTP.Verify(ctx, user.ID); err != nil {
		return nil, "", fmt.Errorf("verify totp: %w", err)
	}

	_ = s.repos.PendingRegistrations.MarkCompleted(ctx, pr.ID)

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &user.ID,
		Action: "registration_completed",
		Metadata: map[string]any{
			"username": user.Username,
			"email":    user.Email,
		},
		IP:        ip,
		CreatedAt: time.Now(),
	})

	sess, sessionToken, err := s.createSession(ctx, user.ID, ip, userAgent, false)
	if err != nil {
		return nil, "", err
	}
	return sess, sessionToken, nil
}

// PrepareInviteSetup generates (or reuses) a TOTP secret for an invite, stores username, returns setup data.
func (s *authService) PrepareInviteSetup(ctx context.Context, inviteToken, username string) (email, totpSecret, totpURI, qrImage string, err error) {
	inv, err := s.repos.Invitations.GetByTokenHash(ctx, hashToken(inviteToken))
	if err != nil || inv.IsExpired() || inv.IsUsed() {
		return "", "", "", "", ErrUnauthorized
	}

	if len(username) < 2 || len(username) > 32 {
		return "", "", "", "", fmt.Errorf("%w: username must be 2–32 characters", ErrInvalidInput)
	}
	// Username uniqueness check.
	if _, uerr := s.repos.Users.GetByUsername(ctx, username); uerr == nil {
		return "", "", "", "", fmt.Errorf("%w: username already taken", ErrConflict)
	}

	var plain string
	if inv.TOTPSecret != "" {
		// Already provisioned on a previous visit — reuse.
		b, err := s.decryptAESGCM(inv.TOTPSecret)
		if err != nil {
			return "", "", "", "", fmt.Errorf("decrypt totp: %w", err)
		}
		plain = string(b)
	} else {
		plain, err = generateTOTPSecret()
		if err != nil {
			return "", "", "", "", fmt.Errorf("generate totp secret: %w", err)
		}
		encrypted, err := s.encryptAESGCM([]byte(plain))
		if err != nil {
			return "", "", "", "", fmt.Errorf("encrypt totp: %w", err)
		}
		if err := s.repos.Invitations.SetTOTPAndUsername(ctx, inv.ID, encrypted, username); err != nil {
			return "", "", "", "", fmt.Errorf("store totp: %w", err)
		}
	}

	issuer := s.settings.TOTPIssuer(ctx).Value
	uri := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(issuer), url.PathEscape(inv.Email),
		plain, url.QueryEscape(issuer),
	)
	return inv.Email, plain, uri, totpQRImage(uri), nil
}

// totpQRImage generates a QR code PNG for the given TOTP URI and returns it
// as a base64-encoded data URI (data:image/png;base64,...). Returns "" on error.
func totpQRImage(uri string) string {
	var buf bytes.Buffer
	png, err := qrcode.Encode(uri, qrcode.Medium, 200)
	if err != nil {
		slog.Warn("totp qr encode", "err", err)
		return ""
	}
	buf.Write(png)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// CompleteInviteSetup finalises an invite-based registration.
func (s *authService) CompleteInviteSetup(ctx context.Context, inviteToken, password, totpCode string, ip, userAgent string) (*model.Session, string, error) {
	if len(password) < 12 {
		return nil, "", fmt.Errorf("%w: password must be at least 12 characters", ErrInvalidInput)
	}

	inv, err := s.repos.Invitations.GetByTokenHash(ctx, hashToken(inviteToken))
	if err != nil || inv.IsExpired() || inv.IsUsed() {
		return nil, "", ErrUnauthorized
	}
	if inv.TOTPSecret == "" || inv.Username == nil {
		return nil, "", fmt.Errorf("%w: setup not initialised — visit the setup page first", ErrInvalidInput)
	}

	plain, err := s.decryptAESGCM(inv.TOTPSecret)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt totp: %w", err)
	}
	if !validateTOTPCode(string(plain), totpCode, time.Now()) {
		return nil, "", ErrInvalidTOTP
	}

	userSvc := newUserService(s.repos, s.cfg)
	user, err := userSvc.Register(ctx, *inv.Username, inv.Email, password, ip, userAgent)
	if err != nil {
		return nil, "", err
	}

	if err := s.repos.TOTP.Create(ctx, user.ID, inv.TOTPSecret); err != nil {
		return nil, "", fmt.Errorf("create totp: %w", err)
	}
	if err := s.repos.TOTP.Verify(ctx, user.ID); err != nil {
		return nil, "", fmt.Errorf("verify totp: %w", err)
	}
	_ = s.repos.Invitations.MarkUsed(ctx, inv.ID, time.Now())

	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &user.ID,
		Action: "registration_completed",
		Metadata: map[string]any{
			"username": user.Username,
			"email":    user.Email,
		},
		IP:        ip,
		CreatedAt: time.Now(),
	})

	sess, sessionToken, err := s.createSession(ctx, user.ID, ip, userAgent, false)
	if err != nil {
		return nil, "", err
	}
	return sess, sessionToken, nil
}

// DeleteAccount hard-deletes the calling user's account after password verification.
func (s *authService) DeleteAccount(ctx context.Context, userID, password string) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if !verifyPassword(password, user.Password) {
		return ErrUnauthorized
	}
	// Log before deletion so the username survives in metadata (user_id becomes NULL after hard delete).
	_ = s.repos.Audit.Log(ctx, &model.AuditEntry{
		ID:     ulid.Make().String(),
		UserID: &userID,
		Action: "account_deleted",
		Metadata: map[string]any{
			"username": user.Username,
			"email":    user.Email,
		},
		CreatedAt: time.Now(),
	})
	return s.repos.Users.HardDelete(ctx, userID)
}
