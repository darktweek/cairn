package service

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
	"strconv"
	"time"

	"github.com/darktweek/cairn/internal/config"
)

// ErrSMTPNotConfigured is returned by send() when the SMTP settings are incomplete.
var ErrSMTPNotConfigured = errors.New("SMTP not configured")

//go:embed templates/password_reset.html
var passwordResetTmpl string

//go:embed templates/welcome.html
var welcomeTmpl string

//go:embed templates/invitation.html
var invitationTmpl string

//go:embed templates/account_setup.html
var accountSetupTmpl string

type EmailService interface {
	SendPasswordReset(ctx context.Context, email, token string) error
	SendWelcome(ctx context.Context, email, username string) error
	// SendInvitation sends an invitation email and returns ErrSMTPNotConfigured
	// when SMTP is not set up, or a real error when delivery fails.
	SendInvitation(ctx context.Context, email, inviteURL string, expiresAt time.Time) error
	SendAccountSetup(ctx context.Context, email, username, setupURL string, expiresAt time.Time) error
	// SendTestEmail sends a plain test message to `to` to verify SMTP settings.
	SendTestEmail(ctx context.Context, to string) error
	// SendCollectionShared notifies a user that a collection was shared with them.
	SendCollectionShared(ctx context.Context, email, sharer, collectionName string) error
}

type emailService struct {
	cfg           *config.Config
	settings      SettingsService
	passwordReset *template.Template
	welcome       *template.Template
	invitation    *template.Template
	accountSetup  *template.Template
}

func newEmailService(cfg *config.Config, settings SettingsService) EmailService {
	return &emailService{
		cfg:           cfg,
		settings:      settings,
		passwordReset: template.Must(template.New("password_reset").Parse(passwordResetTmpl)),
		welcome:       template.Must(template.New("welcome").Parse(welcomeTmpl)),
		invitation:    template.Must(template.New("invitation").Parse(invitationTmpl)),
		accountSetup:  template.Must(template.New("account_setup").Parse(accountSetupTmpl)),
	}
}

func (s *emailService) SendPasswordReset(ctx context.Context, email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.BaseURL, token)
	siteName := s.settings.SiteName(ctx)

	data := struct {
		SiteName string
		Username string
		ResetURL string
	}{
		SiteName: siteName,
		Username: email,
		ResetURL: resetURL,
	}

	body, err := renderTemplate(s.passwordReset, data)
	if err != nil {
		slog.Error("render password reset email", "err", err)
		return nil
	}

	if err := s.send(ctx, email, "Réinitialisation de mot de passe — "+siteName, body); err != nil && !errors.Is(err, ErrSMTPNotConfigured) {
		slog.Error("send password reset email", "err", err)
	}
	return nil
}

func (s *emailService) SendWelcome(ctx context.Context, email, username string) error {
	siteName := s.settings.SiteName(ctx)
	data := struct {
		SiteName string
		Username string
		BaseURL  string
	}{
		SiteName: siteName,
		Username: username,
		BaseURL:  s.cfg.BaseURL,
	}

	body, err := renderTemplate(s.welcome, data)
	if err != nil {
		slog.Error("render welcome email", "err", err)
		return nil
	}

	if err := s.send(ctx, email, "Bienvenue sur "+siteName, body); err != nil && !errors.Is(err, ErrSMTPNotConfigured) {
		slog.Error("send welcome email", "err", err)
	}
	return nil
}

// SendInvitation returns nil on success, ErrSMTPNotConfigured when SMTP is not
// set up, or an error when SMTP is configured but delivery fails.
func (s *emailService) SendInvitation(ctx context.Context, email, inviteURL string, expiresAt time.Time) error {
	siteName := s.settings.SiteName(ctx)
	data := struct {
		SiteName  string
		InviteURL string
		ExpiresAt string
	}{
		SiteName:  siteName,
		InviteURL: inviteURL,
		ExpiresAt: expiresAt.Format("02/01/2006 à 15:04"),
	}
	body, err := renderTemplate(s.invitation, data)
	if err != nil {
		return fmt.Errorf("render invitation email: %w", err)
	}
	return s.send(ctx, email, "Invitation à rejoindre "+siteName, body)
}

func (s *emailService) SendAccountSetup(ctx context.Context, email, username, setupURL string, expiresAt time.Time) error {
	siteName := s.settings.SiteName(ctx)
	data := struct {
		SiteName  string
		Username  string
		SetupURL  string
		ExpiresAt string
	}{
		SiteName:  siteName,
		Username:  username,
		SetupURL:  setupURL,
		ExpiresAt: expiresAt.Format("02/01/2006 à 15:04"),
	}
	body, err := renderTemplate(s.accountSetup, data)
	if err != nil {
		slog.Error("render account setup email", "err", err)
		return nil
	}
	if err := s.send(ctx, email, "Finalisez votre compte "+siteName, body); err != nil && !errors.Is(err, ErrSMTPNotConfigured) {
		slog.Error("send account setup email", "err", err)
	}
	return nil
}

func (s *emailService) SendTestEmail(ctx context.Context, to string) error {
	siteName := s.settings.SiteName(ctx)
	body := `<!DOCTYPE html><html><body style="font-family:sans-serif;max-width:480px;margin:40px auto;padding:20px">` +
		`<h2 style="color:#1a1a2e">Test SMTP — ` + template.HTMLEscapeString(siteName) + `</h2>` +
		`<p>Si vous voyez cet email, votre configuration SMTP fonctionne correctement.</p>` +
		`<p style="color:#888;font-size:12px">Envoyé depuis l'interface d'administration ` + template.HTMLEscapeString(siteName) + `.</p>` +
		`</body></html>`
	if err := s.send(ctx, to, "Test SMTP — "+siteName, body); err != nil {
		if errors.Is(err, ErrSMTPNotConfigured) {
			return fmt.Errorf("%w: SMTP non configuré", ErrInvalidInput)
		}
		return err
	}
	return nil
}

func (s *emailService) SendCollectionShared(ctx context.Context, email, sharer, collectionName string) error {
	siteName := s.settings.SiteName(ctx)
	body := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:sans-serif;max-width:480px;margin:40px auto;padding:20px">`+
		`<h2 style="color:#1a1a2e">Une collection a été partagée avec vous — %s</h2>`+
		`<p><strong>%s</strong> a partagé la collection « <strong>%s</strong> » avec vous.</p>`+
		`<p><a href="%s" style="color:#4f46e5">Ouvrir %s</a></p>`+
		`<p style="color:#888;font-size:12px">Vous recevez cet email car votre compte %s a reçu un accès partagé.</p>`+
		`</body></html>`,
		template.HTMLEscapeString(siteName),
		template.HTMLEscapeString(sharer), template.HTMLEscapeString(collectionName), s.cfg.BaseURL,
		template.HTMLEscapeString(siteName), template.HTMLEscapeString(siteName))
	if err := s.send(ctx, email, "Collection partagée — "+siteName, body); err != nil && !errors.Is(err, ErrSMTPNotConfigured) {
		slog.Error("send collection shared email", "err", err)
	}
	return nil
}

func (s *emailService) send(ctx context.Context, to, subject, htmlBody string) error {
	cfg := s.settings.SMTP(ctx)
	if !cfg.Configured() {
		slog.Warn("email not sent: SMTP is not configured", "to", to, "subject", subject)
		return ErrSMTPNotConfigured
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	tlsCfg := &tls.Config{ServerName: cfg.Host}

	// Port 465 = SMTPS (implicit TLS from the start).
	// Port 587/25 with TLS = STARTTLS upgrade after plain handshake.
	implicitTLS := cfg.TLS && cfg.Port == 465

	var conn net.Conn
	if implicitTLS {
		tlsConn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		conn = tlsConn
	} else {
		var err error
		conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return fmt.Errorf("smtp dial: %w", err)
		}
	}

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if cfg.TLS && !implicitTLS {
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if cfg.User != "" {
		auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := c.Mail(cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	defer w.Close()

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.From, to, subject, htmlBody,
	)

	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return nil
}

func renderTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
