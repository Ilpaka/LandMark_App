package app

import (
	"context"
	"time"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

// RateLimitInfo describes sliding-window state after one Allow evaluation (for X-RateLimit-* headers).
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetUnix int64
}

func (s *Service) rateConsume(ctx context.Context, key string, window time.Duration, limit int) (allowed bool, info RateLimitInfo, err error) {
	info.Limit = limit
	ok, rem, reset, e := s.Redis.Allow(ctx, key, window, limit)
	if e != nil {
		return false, info, domain.ErrServiceUnavailable
	}
	if rem < 0 {
		rem = 0
	}
	info.Remaining = rem
	info.ResetUnix = reset.Unix()
	return ok, info, nil
}

// RateRegister consumes one registration attempt token for the client IP.
func (s *Service) RateRegister(ctx context.Context, ip string) (RateLimitInfo, error) {
	ok, info, err := s.rateConsume(ctx, "rl:register:"+ip, rateRegisterWindow, rateRegisterLimit)
	if err != nil {
		return info, err
	}
	if !ok {
		return info, domain.ErrRateLimited
	}
	return info, nil
}

func mergeRateLimit(a, b RateLimitInfo) RateLimitInfo {
	rem := a.Remaining
	if b.Remaining < rem {
		rem = b.Remaining
	}
	rst := a.ResetUnix
	if b.ResetUnix > rst {
		rst = b.ResetUnix
	}
	return RateLimitInfo{Limit: a.Limit, Remaining: rem, ResetUnix: rst}
}
