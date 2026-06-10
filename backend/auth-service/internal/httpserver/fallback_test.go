package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// BUG-03: unknown route must return the JSON error envelope,
// not gin's default plain-text "404 page not found".
func TestFallbackUnknownRouteReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterFallbackHandlers(r)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/unknown", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json (body: %q)", ct, w.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v (body: %q)", err, w.Body.String())
	}
	if body["error"] != "not found" {
		t.Fatalf(`body["error"] = %q, want "not found"`, body["error"])
	}
}

func TestFallbackUnknownMethodReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterFallbackHandlers(r)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/healthz", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405 (body: %q)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json (body: %q)", ct, w.Body.String())
	}
}
