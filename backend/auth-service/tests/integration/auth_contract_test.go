//go:build integration

package integration

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/httpserver"
	"github.com/ilpaka/landmark_app/backend/auth-service/tests/testutil"
)

func loadAndValidateSpec(t *testing.T, openAPIPath string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(openAPIPath)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := doc.Validate(ctx); err != nil {
		t.Fatalf("openapi validate: %v", err)
	}
	return doc
}

func validateResponse(t *testing.T, doc *openapi3.T, req *http.Request, resp *http.Response, body []byte) {
	t.Helper()
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("find route %s %s: %v", req.Method, req.URL.Path, err)
	}
	opts := &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}
	rvi := &openapi3filter.RequestValidationInput{
		Request:    req,
		Route:      route,
		PathParams: pathParams,
		Options:    opts,
	}
	ctx := context.Background()
	if err := openapi3filter.ValidateRequest(ctx, rvi); err != nil {
		t.Fatalf("validate request: %v", err)
	}
	rv := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: rvi,
		Status:                 resp.StatusCode,
		Header:                 resp.Header.Clone(),
		Body:                   io.NopCloser(bytes.NewReader(body)),
		Options:                opts,
	}
	if err := openapi3filter.ValidateResponse(ctx, rv); err != nil {
		t.Fatalf("validate response: %v", err)
	}
}

// validateResponseOnly checks the HTTP response against the OpenAPI spec for the route (no request body validation).
// Use when the request body is intentionally invalid but the route is still identifiable.
func validateResponseOnly(t *testing.T, doc *openapi3.T, req *http.Request, resp *http.Response, body []byte) {
	t.Helper()
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("find route %s %s: %v", req.Method, req.URL.Path, err)
	}
	opts := &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}
	rvi := &openapi3filter.RequestValidationInput{
		Request:    req,
		Route:      route,
		PathParams: pathParams,
		Options:    opts,
	}
	ctx := context.Background()
	rv := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: rvi,
		Status:                 resp.StatusCode,
		Header:                 resp.Header.Clone(),
		Body:                   io.NopCloser(bytes.NewReader(body)),
		Options:                opts,
	}
	if err := openapi3filter.ValidateResponse(ctx, rv); err != nil {
		t.Fatalf("validate response: %v", err)
	}
}

func TestAuthContract_RegisterAndHealth(t *testing.T) {
	t.Setenv("GIN_MODE", "release")

	deps := testutil.RequireAuthDocker(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.HTTPAddr = "127.0.0.1:0"

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	if err != nil {
		t.Fatal(err)
	}

	openAPIPath := filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml")
	doc := loadAndValidateSpec(t, openAPIPath)

	{
		req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/healthz", nil)
		w := httptest.NewRecorder()
		stack.Engine.ServeHTTP(w, req)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("healthz: %d %s", resp.StatusCode, body)
		}
		if resp.Header.Get("X-Request-ID") == "" {
			t.Fatal("healthz: missing X-Request-ID")
		}
		validateResponse(t, doc, req, resp, body)
	}
}
