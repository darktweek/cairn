package service

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
	"time"

	"github.com/darktweek/cairn/internal/config"
)

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
	SendInvitation(ctx context.Context, email, inviteURL string, expiresAt time.Time) error
	SendAccountSetup(ctx context.Context, email, username, setupURL string, expiresAt time.Time) error
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

	data := struct {
		Username string
		ResetURL string
	}{
		Username: email,
		ResetURL: resetURL,
	}

	body, err := renderTemplate(s.passwordReset, data)
	if err != nil {
		slog.Error("render password reset email", "err", err)
		return nil
	}

	if err := s.send(ctx, email, "Réinitialisation de mot de passe — Cairn", body); err != nil {
		slog.Error("send password reset email", "err", err)
	}
	return nil
}

func (s *emailService) SendWelcome(ctx context.Context, email, username string) error {
	data := struct {
		Username string
		BaseURL  string
	}{
		Username: username,
		BaseURL:  s.cfg.BaseURL,
	}

	body, err := renderTemplate(s.welcome, data)
	if err != nil {
		slog.Error("render welcome email", "err", err)
		return nil
	}

	if err := s.send(ctx, email, "Bienvenue sur Cairn", body); err != nil {
		slog.Error("send welcome email", "err", err)
	}
	return nil
}

func (s *emailService) SendInvitation(ctx context.Context, email, inviteURL string, expiresAt time.Time) error {
	data := struct {
		InviteURL string
		ExpiresAt string
	}{
		InviteURL: inviteURL,
		ExpiresAt: expiresAt.Format("02/01/2006 à 15:04"),
	}
	body, err := renderTemplate(s.invitation, data)
	if err != nil {
		slog.Error("render invitation email", "err", err)
		return nil
	}
	if err := s.send(ctx, email, "Invitation à rejoindre Cairn", body); err != nil {
		slog.Error("send invitation email", "err", err)
	}
	return nil
}

func (s *emailService) SendAccountSetup(ctx context.Context, email, username, setupURL string, expiresAt time.Time) error {
	data := struct {
		Username  string
		SetupURL  string
		ExpiresAt string
	}{
		Username:  username,
		SetupURL:  setupURL,
		ExpiresAt: expiresAt.Format("02/01/2006 à 15:04"),
	}
	body, err := renderTemplate(s.accountSetup, data)
	if err != nil {
		slog.Error("render account setup email", "err", err)
		return nil
	}
	if err := s.send(ctx, email, "Finalisez votre compte Cairn", body); err != nil {
		slog.Error("send account setup email", "err", err)
	}
	return nil
}

func (s *emailService) send(ctx context.Context, to, subject, htmlBody string) error {
	cfg := s.settings.SMTP(ctx)
	if !cfg.Configured() {
		slog.Warn("email not sent: SMTP is not configured", "to", to, "subject", subject)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if cfg.TLS {
		tlsCfg := &tls.Config{ServerName: cfg.Host}
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
