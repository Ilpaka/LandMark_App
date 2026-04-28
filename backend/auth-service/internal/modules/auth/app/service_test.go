package app

import (
	"errors"
	"testing"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()
	email, norm, err := normalizeEmail("  User@Example.COM ")
	if err != nil {
		t.Fatal(err)
	}
	if norm != "user@example.com" {
		t.Fatalf("norm: %q", norm)
	}
	if email == "" {
		t.Fatal("empty email")
	}
	if _, _, err := normalizeEmail("not-an-email"); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
}

func TestDomainFromEmail(t *testing.T) {
	t.Parallel()
	if got := domainFromEmail("user@Example.com"); got != "Example.com" {
		t.Fatalf("got %q", got)
	}
	if got := domainFromEmail("nope"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestMergeStr(t *testing.T) {
	t.Parallel()
	if got := mergeStr(ptrStr("old"), ptrStr("new")); got == nil || *got != "new" {
		t.Fatalf("mergeStr new: %v", got)
	}
	if got := mergeStr(ptrStr("old"), ptrStr("")); got == nil || *got != "old" {
		t.Fatalf("mergeStr empty neu: %v", got)
	}
	if got := mergeStr(ptrStr("old"), nil); got == nil || *got != "old" {
		t.Fatalf("mergeStr nil neu: %v", got)
	}
}

func ptrStr(s string) *string { return &s }
