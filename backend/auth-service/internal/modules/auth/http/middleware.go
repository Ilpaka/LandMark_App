package authhttp

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

type ctxKey string

const (
	ctxClaims ctxKey = "auth_claims"
)

// AuthMiddleware validates access JWT and optional blacklist.
func AuthMiddleware(jwt ports.JWTSigner, rdb *redisx.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
			return
		}
		raw := strings.TrimSpace(h[7:])
		cl, err := jwt.ParseAccess(c.Request.Context(), raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		if time.Now().After(cl.Expires) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_expired"})
			return
		}
		if rdb != nil {
			blocked, err := rdb.Has(c.Request.Context(), cl.JTI)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
				return
			}
			if blocked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_revoked"})
				return
			}
		}
		c.Set(string(ctxClaims), cl)
		c.Next()
	}
}

func claimsFromCtx(c *gin.Context) (ports.AccessClaims, bool) {
	v, ok := c.Get(string(ctxClaims))
	if !ok {
		return ports.AccessClaims{}, false
	}
	cl, ok := v.(ports.AccessClaims)
	return cl, ok
}

// RequireAdmin ensures role admin on JWT claims.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl, ok := claimsFromCtx(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if cl.Role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
