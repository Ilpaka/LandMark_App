package audit

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

func TestMaskEvent_IPAndMeta(t *testing.T) {
	ip := "192.168.1.99"
	ua := "Mozilla/5.0 (very long) " + string(make([]byte, 300))
	ev := ports.AuditEvent{
		ID:        uuid.New(),
		UserID:    ptrUUID(uuid.MustParse("11111111-1111-1111-1111-111111111111")),
		EventType: "login_failed",
		IP:        &ip,
		UserAgent: &ua,
		Meta: map[string]any{
			"email":      "secret@example.com",
			"session_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			"nested":     map[string]any{"token": "x"},
		},
		CreatedAt: time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}
	m := MaskEvent(ev)
	if m.IP != nil {
		t.Fatal("expected ip cleared")
	}
	if m.Meta["ip_class"] != "192.168.*.*" {
		t.Fatalf("ip_class: %v", m.Meta["ip_class"])
	}
	if m.Meta["email"] != "[redacted]" {
		t.Fatalf("email: %v", m.Meta["email"])
	}
	if m.Meta["session_id"] != "aaaaaaaa…" {
		t.Fatalf("session_id: %v", m.Meta["session_id"])
	}
	n := m.Meta["nested"].(map[string]any)
	if n["token"] != "[redacted]" {
		t.Fatalf("nested token: %v", n["token"])
	}
	if len(*m.UserAgent) > userAgentMaxRunes+4 {
		t.Fatalf("user agent not truncated: len=%d", len(*m.UserAgent))
	}
}

func TestEventToMapRoundTrip(t *testing.T) {
	ev := MaskEvent(ports.AuditEvent{
		ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		EventType: "email_verified",
		Meta:      map[string]any{"k": "v"},
		CreatedAt: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
	})
	m, err := EventToMap(ev)
	if err != nil {
		t.Fatal(err)
	}
	got, err := MapToEvent(m)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != ev.ID || got.EventType != ev.EventType {
		t.Fatalf("mismatch %+v vs %+v", got, ev)
	}
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	return &u
}
