package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware_Generated(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/t", func(c *gin.Context) {
		v, _ := c.Get(ginCtxRequestID)
		s, _ := v.(string)
		if s == "" {
			c.Status(500)
			return
		}
		c.Status(204)
	})
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("code %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID")
	}
}

func TestRequestIDMiddleware_FromHeader(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/t", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set("X-Request-ID", "client-rid-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-ID"); got != "client-rid-1" {
		t.Fatalf("header %q", got)
	}
}

func TestPrometheusHTTPMiddleware(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(PrometheusHTTPMiddleware())
	r.GET("/ok", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("code %d", w.Code)
	}
}
