package authhttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrBadRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
	case errors.Is(err, domain.ErrNicknameUnavailable):
		c.JSON(http.StatusConflict, gin.H{"error": "nickname_unavailable"})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
	case errors.Is(err, domain.ErrInvalidOTP), errors.Is(err, domain.ErrTooManyAttempts):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
	case errors.Is(err, domain.ErrAccountBlocked):
		c.JSON(http.StatusForbidden, gin.H{"error": "account_blocked"})
	case errors.Is(err, domain.ErrAccountDeleted):
		c.JSON(http.StatusForbidden, gin.H{"error": "account_deleted"})
	case errors.Is(err, domain.ErrEmailNotVerified):
		c.JSON(http.StatusForbidden, gin.H{"error": "email_not_verified"})
	case errors.Is(err, domain.ErrInvalidRefresh):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
	case errors.Is(err, domain.ErrRefreshReuse):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_reuse"})
	case errors.Is(err, domain.ErrRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited"})
	case errors.Is(err, domain.ErrServiceUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}
