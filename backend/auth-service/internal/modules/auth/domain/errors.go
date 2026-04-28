package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidOTP          = errors.New("invalid or expired code")
	ErrTooManyAttempts     = errors.New("too many attempts")
	ErrAccountBlocked      = errors.New("account blocked")
	ErrAccountDeleted      = errors.New("account deleted")
	ErrEmailNotVerified    = errors.New("email not verified")
	ErrInvalidRefresh      = errors.New("invalid refresh token")
	ErrRefreshReuse        = errors.New("refresh token reuse detected")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrRateLimited         = errors.New("rate limited")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrForbidden           = errors.New("forbidden")
	ErrBadRequest          = errors.New("bad request")
	ErrNicknameUnavailable = errors.New("nickname unavailable")
)
