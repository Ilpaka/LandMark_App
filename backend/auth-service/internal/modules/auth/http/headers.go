package authhttp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/app"
)

// WriteRateLimitHeaders sets standard-ish X-RateLimit-* headers (Reset = Unix seconds).
func WriteRateLimitHeaders(c *gin.Context, rl app.RateLimitInfo) {
	if rl.Limit <= 0 {
		return
	}
	c.Header("X-RateLimit-Limit", strconv.Itoa(rl.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.Remaining))
	if rl.ResetUnix > 0 {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(rl.ResetUnix, 10))
	}
}
