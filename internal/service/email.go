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

type EmailService interface {
	SendPasswordReset(ctx context.Context, email, token string) error
	SendWelcome(ctx context.Context, email, username string) error
	SendInvitation(ctx context.Context, email, inviteURL string, expiresAt time.Time) error
}

type emailService struct {
	cfg           *config.Config
	passwordReset *template.Template
	welcome       *template.Template
	invitation    *template.Template
}

func newEmailService(cfg *config.Config) EmailService {
	return &emailService{
		cfg:           cfg,
		passwordReset: template.Must(template.New("password_reset").Parse(passwordResetTmpl)),
		welcome:       template.Must(template.New("welcome").Parse(welcomeTmpl)),
		invitation:    template.Must(template.New("invitation").Parse(invitationTmpl)),
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

	if err := s.send(email, "Réinitialisation de mot de passe — Cairn", body); err != nil {
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

	if err := s.send(email, "Bienvenue sur Cairn", body); err != nil {
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
	if err := s.send(email, "Invitation à rejoindre Cairn", body); err != nil {
		slog.Error("send invitation email", "err", err)
	}
	return nil
}

func (s *emailService) send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	c, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if s.cfg.SMTPTLS {
		tlsCfg := &tls.Config{ServerName: s.cfg.SMTPHost}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := c.Mail(s.cfg.SMTPFrom); err != nil {
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
		s.cfg.SMTPFrom, to, subject, htmlBody,
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
