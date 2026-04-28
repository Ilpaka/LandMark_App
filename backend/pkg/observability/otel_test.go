package observability

import (
	"context"
	"testing"
	"time"
)

func TestNormalizeOTLPEndpoint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"  ", ""},
		{"http://127.0.0.1:4318", "http://127.0.0.1:4318/v1/traces"},
		{"http://127.0.0.1:4318/", "http://127.0.0.1:4318/v1/traces"},
		{"http://127.0.0.1:4318/v1/traces", "http://127.0.0.1:4318/v1/traces"},
		{"https://collector.example/v1/traces", "https://collector.example/v1/traces"},
		{"not-a-url", "not-a-url"},
	}
	for _, tc := range cases {
		if got := normalizeOTLPEndpoint(tc.in); got != tc.want {
			t.Fatalf("normalizeOTLPEndpoint(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSetupOTel_EmptyEndpointShutdown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shutdown, err := SetupOTel(ctx, "test-svc", "")
	if err != nil {
		t.Fatal(err)
	}
	sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := shutdown(sctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
