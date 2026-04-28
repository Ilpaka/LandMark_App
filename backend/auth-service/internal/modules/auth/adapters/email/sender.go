package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
)

// LogSender logs OTPs (development-friendly).
type LogSender struct {
	Log *slog.Logger
}

func (l LogSender) SendVerificationOTP(ctx context.Context, toEmail, code string) error {
	_ = ctx
	l.Log.Info("verification otp", "to", toEmail, "code", code)
	return nil
}

func (l LogSender) SendPasswordResetOTP(ctx context.Context, toEmail, code string) error {
	_ = ctx
	l.Log.Info("password reset otp", "to", toEmail, "code", code)
	return nil
}

// SMTPConfig for MailHog / real SMTP.
type SMTPConfig struct {
	Addr string
	From string
}

// SMTPSender sends plain-text emails.
type SMTPSender struct {
	Cfg SMTPConfig
	Log *slog.Logger
}

func (s SMTPSender) SendVerificationOTP(ctx context.Context, toEmail, code string) error {
	return s.send(ctx, toEmail, "Verify your email", fmt.Sprintf("Your verification code is: %s", code))
}

func (s SMTPSender) SendPasswordResetOTP(ctx context.Context, toEmail, code string) error {
	return s.send(ctx, toEmail, "Reset your password", fmt.Sprintf("Your reset code is: %s", code))
}

func (s SMTPSender) send(ctx context.Context, to, subject, body string) error {
	if s.Cfg.Addr == "" {
		if s.Log != nil {
			s.Log.Info("email (no smtp)", "to", to, "subject", subject, "body", body)
		}
		return nil
	}
	from := s.Cfg.From
	if from == "" {
		from = "auth@localhost"
	}
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", from, to, subject, body))
	host := s.Cfg.Addr
	_ = ctx
	return smtp.SendMail(host, nil, from, []string{to}, msg)
}

func NewFromEnv(log *slog.Logger) SMTPSender {
	addr := os.Getenv("SMTP_ADDR") // e.g. localhost:1025 for MailHog
	from := os.Getenv("SMTP_FROM")
	return SMTPSender{Cfg: SMTPConfig{Addr: addr, From: from}, Log: log}
}
