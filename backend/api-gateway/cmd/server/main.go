package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/api-gateway/internal/middleware"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type upstream struct {
	prefix string
	url    string
}

var upstreams = []upstream{
	{"/v1/auth", ""},
	{"/v1/admin", ""},
	{"/v1/profile", ""},
	{"/v1/places", ""},
	{"/v1/trips", ""},
	{"/v1/journal", ""},
	{"/v1/media", ""},
	{"/v1/favorites", ""},
	{"/v1/notifications", ""},
	{"/v1/moderation", ""},
}

// Routes where auth is optional (public reads + auth endpoints themselves)
var authOptionalPrefixes = []string{
	"/v1/auth/register",
	"/v1/auth/verify-email",
	"/v1/auth/login",
	"/v1/auth/refresh",
	"/v1/auth/forgot-password",
	"/v1/auth/reset-password",
	"/v1/auth/jwks",
	"/v1/auth/phone",
}

func isAuthOptional(method, path string) bool {
	for _, p := range authOptionalPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	// Public place reads
	if method == http.MethodGet && strings.HasPrefix(path, "/v1/places") {
		return true
	}
	return false
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	// Wire upstream URLs from env
	urlMap := map[string]string{
		"/v1/auth":          getenv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		"/v1/admin":         getenv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		"/v1/profile":       getenv("PROFILE_SERVICE_URL", "http://profile-service:8080"),
		"/v1/places":        getenv("PLACES_SERVICE_URL", "http://places-service:8080"),
		"/v1/trips":         getenv("TRIPS_SERVICE_URL", "http://trips-service:8080"),
		"/v1/journal":       getenv("JOURNAL_SERVICE_URL", "http://journal-service:8080"),
		"/v1/media":         getenv("MEDIA_SERVICE_URL", "http://media-service:8080"),
		"/v1/favorites":     getenv("FAVORITES_SERVICE_URL", "http://favorites-service:8080"),
		"/v1/notifications": getenv("NOTIFICATIONS_SERVICE_URL", "http://notifications-service:8080"),
		"/v1/moderation":    getenv("MODERATION_SERVICE_URL", "http://moderation-service:8080"),
	}
	for i, u := range upstreams {
		upstreams[i].url = urlMap[u.prefix]
	}

	jwtMW := middleware.NewJWTMiddleware(
		getenv("AUTH_JWKS_URL", "http://auth-service:8080/v1/auth/jwks"),
		getenv("JWT_ISSUER", "landmark.app"),
		getenv("JWT_AUDIENCE", "landmark-clients"),
	)

	corsOrigins := getenv("CORS_ORIGINS", "*")
	client := &http.Client{Timeout: 30 * time.Second}

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Block all internal paths
		if strings.Contains(path, "/internal/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Request ID
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// CORS
		c.Header("Access-Control-Allow-Origin", corsOrigins)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-Id")
		c.Header("X-Request-Id", reqID)

		if method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}

		// Strip any client-supplied internal headers
		c.Request.Header.Del("X-User-Id")
		c.Request.Header.Del("X-Role")
		c.Request.Header.Del("X-Session-Id")
		c.Request.Header.Del("X-Internal-Key")

		// JWT auth
		if isAuthOptional(method, path) {
			jwtMW.Optional()(c)
		} else {
			jwtMW.Required()(c)
			if c.IsAborted() {
				return
			}
		}

		// Find upstream
		var targetURL string
		for _, u := range upstreams {
			if strings.HasPrefix(path, u.prefix) {
				targetURL = u.url + path
				if c.Request.URL.RawQuery != "" {
					targetURL += "?" + c.Request.URL.RawQuery
				}
				break
			}
		}

		if targetURL == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Proxy
		req, err := http.NewRequestWithContext(c.Request.Context(), method, targetURL, c.Request.Body)
		if err != nil {
			log.Error("proxy build", "err", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "bad gateway"})
			return
		}
		for k, vv := range c.Request.Header {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
		req.Header.Set("X-Request-Id", reqID)

		resp, err := client.Do(req)
		if err != nil {
			log.Error("proxy upstream", "url", targetURL, "err", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "bad gateway"})
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				c.Header(k, v)
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	})

	addr := getenv("HTTP_ADDR", ":8080")
	log.Info("api-gateway starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Error("server fatal", "err", err)
		os.Exit(1)
	}
}
