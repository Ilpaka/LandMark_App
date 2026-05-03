package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
}

type JWTMiddleware struct {
	Cache    *JWKSCache
	Issuer   string
	Audience string
}

func NewJWTMiddleware(jwksURL, issuer, audience string) *JWTMiddleware {
	return &JWTMiddleware{
		Cache:    NewJWKSCache(jwksURL, time.Hour),
		Issuer:   issuer,
		Audience: audience,
	}
}

func (m *JWTMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extract(c)
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		m.inject(c, claims)
		c.Next()
	}
}

func (m *JWTMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extract(c)
		if err == nil && claims != nil {
			m.inject(c, claims)
		}
		// Always continue regardless of auth state
		c.Next()
	}
}

func (m *JWTMiddleware) extract(c *gin.Context) (*Claims, error) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, nil
	}
	raw := h[7:]

	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		kid, _ := t.Header["kid"].(string)
		return m.Cache.GetKey(context.Background(), kid)
	}, jwt.WithIssuer(m.Issuer), jwt.WithAudience(m.Audience), jwt.WithExpirationRequired())

	if err != nil || !token.Valid {
		return nil, err
	}
	return &claims, nil
}

func (m *JWTMiddleware) inject(c *gin.Context, claims *Claims) {
	c.Request.Header.Del("X-User-Id")
	c.Request.Header.Del("X-Role")
	c.Request.Header.Del("X-Session-Id")
	c.Request.Header.Set("X-User-Id", claims.Subject)
	c.Request.Header.Set("X-Role", claims.Role)
	if claims.SessionID != "" {
		c.Request.Header.Set("X-Session-Id", claims.SessionID)
	}
}
