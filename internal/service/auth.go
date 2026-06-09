package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
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
	"github.com/darktweek/cairn/internal/repository"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/argon2"
)

const (
	sessionLifetime    = 30 * 24 * time.Hour
	bookmarkletLifetime = 90 * 24 * time.Hour
	resetTokenLifetime  = 1 * time.Hour
	argon2Time         = 1
	argon2Memory       = 64 * 1024 // 64 MB
	argon2Threads      = 4
	argon2KeyLen       = 32
	argon2SaltLen      = 16
)

type AuthService interface {
	Login(ctx context.Context, email, password, totpCode, ip, userAgent string) (*model.Session, string, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutForUser(ctx context.Context, sessionID, userID string) error
	LogoutAll(ctx context.Context, userID string) error
	ValidateSession(ctx context.Context, token string) (*model.User, *model.Session, error)
	CreateBookmarkletSession(ctx context.Context, userID, ip, userAgent string) (string, error)
	BeginTOTP(ctx context.Context, userID string) (secret, qrCodeURL string, err error)
	ConfirmTOTP(ctx context.Context, userID, code string) error
	ValidateTOTP(ctx context.Context, userID, code string) (bool, error)
	DisableTOTP(ctx context.Context, userID, password string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type resetEntry struct {
	userID    string
	expiresAt time.Time
}

type authService struct {
	repos       *repository.Repositories
	cfg         *config.Config
	resetTokens sync.Map // tokenHash → resetEntry
}

func newAuthService(repos *repository.Repositories, cfg *config.Config) AuthService {
	return &authService{repos: repos, cfg: cfg}
}

// Login authenticates a user. totpCode is optional — if TOTP is enabled and empty, returns ErrTOTPRequired.
func (s *authService) Login(ctx context.Context, email, password, totpCode, ip, userAgent string) (*model.Session, string, error) {
	user, err := s.repos.Users.GetByEmail(ctx, email)
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

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	return s.repos.Sessions.DeleteByID(ctx, sessionID)
}

// LogoutForUser revokes a session only if it belongs to the given user — prevents IDOR.
func (svc *authService) LogoutForUser(ctx context.Context, sessionID, userID string) error {
	sessions, err := svc.repos.Sessions.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		if sess.ID == sessionID {
			return svc.repos.Sessions.DeleteByID(ctx, sessionID)
		}
	}
	return ErrNotFound
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	return s.repos.Sessions.DeleteByUserID(ctx, userID)
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

	return user, sess, nil
}

func (s *authService) BeginTOTP(ctx context.Context, userID string) (string, string, error) {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return "", "", ErrNotFound
	}

	secret, err := generateTOTPSecret()
	if err != nil {
		return "", "", fmt.Errorf("generate totp secret: %w", err)
	}

	encrypted, err := s.encryptAESGCM([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("encrypt totp secret: %w", err)
	}

	if err := s.repos.TOTP.Create(ctx, userID, encrypted); err != nil {
		return "", "", err
	}

	qrURL := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.PathEscape(s.cfg.TOTPIssuer),
		url.PathEscape(user.Email),
		secret,
		url.QueryEscape(s.cfg.TOTPIssuer),
	)

	return secret, qrURL, nil
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

	return s.repos.TOTP.Verify(ctx, userID)
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

	return nil
}

// createSession generates a cryptographically random token, stores its SHA-256 hash.
func (s *authService) createSession(ctx context.Context, userID, ip, userAgent string, isBookmarklet bool) (*model.Session, string, error) {
	token, err := generateRawToken()
	if err != nil {
		return nil, "", err
	}

	lifetime := sessionLifetime
	if isBookmarklet {
		lifetime = bookmarkletLifetime
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
	key := deriveKey(s.cfg.SessionSecret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

// decryptAESGCM decrypts a base64-encoded AES-GCM ciphertext.
func (s *authService) decryptAESGCM(encoded string) ([]byte, error) {
	data, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	key := deriveKey(s.cfg.SessionSecret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// deriveKey produces a 32-byte AES key from the session secret via SHA-256.
func deriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}
