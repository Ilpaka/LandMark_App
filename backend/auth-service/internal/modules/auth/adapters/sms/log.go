package sms

import (
	"context"
	"log/slog"
)

// LogSender logs SMS content in development (code is not included in production logs; still avoid in prod best-effort).
type LogSender struct {
	Log *slog.Logger
}

func (l LogSender) SendOTP(_ context.Context, phoneE164, code string) error {
	if l.Log != nil {
		// "code" key matches e2e regexp `code=(\d{6})` (text log handler).
		l.Log.Info("sms otp", "to", phoneE164, "code", code)
	}
	return nil
}
