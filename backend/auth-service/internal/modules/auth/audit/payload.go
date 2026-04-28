package audit

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// EventToMap serializes a masked audit event for the outbox payload.
func EventToMap(e ports.AuditEvent) (map[string]any, error) {
	meta := e.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaCopy := make(map[string]any, len(meta))
	for k, v := range meta {
		metaCopy[k] = v
	}
	m := map[string]any{
		"id":         e.ID.String(),
		"event_type": e.EventType,
		"meta":       metaCopy,
		"created_at": e.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if e.UserID != nil {
		m["user_id"] = e.UserID.String()
	}
	if e.IP != nil {
		m["ip"] = *e.IP
	}
	if e.UserAgent != nil {
		m["user_agent"] = *e.UserAgent
	}
	return m, nil
}

// MapToEvent decodes outbox payload written by EventToMap.
func MapToEvent(payload map[string]any) (ports.AuditEvent, error) {
	var e ports.AuditEvent
	raw, err := json.Marshal(payload)
	if err != nil {
		return e, err
	}
	var wire struct {
		ID        string         `json:"id"`
		UserID    *string        `json:"user_id"`
		EventType string         `json:"event_type"`
		IP        *string        `json:"ip"`
		UserAgent *string        `json:"user_agent"`
		Meta      map[string]any `json:"meta"`
		CreatedAt string         `json:"created_at"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return e, err
	}
	id, err := uuid.Parse(wire.ID)
	if err != nil {
		return e, fmt.Errorf("id: %w", err)
	}
	e.ID = id
	e.EventType = wire.EventType
	if wire.UserID != nil {
		u, err := uuid.Parse(*wire.UserID)
		if err != nil {
			return e, fmt.Errorf("user_id: %w", err)
		}
		e.UserID = &u
	}
	e.IP = wire.IP
	e.UserAgent = wire.UserAgent
	if wire.Meta != nil {
		e.Meta = wire.Meta
	} else {
		e.Meta = map[string]any{}
	}
	if wire.CreatedAt == "" {
		return e, fmt.Errorf("created_at required")
	}
	ts, err := time.Parse(time.RFC3339Nano, wire.CreatedAt)
	if err != nil {
		ts, err = time.Parse(time.RFC3339, wire.CreatedAt)
		if err != nil {
			return e, fmt.Errorf("created_at: %w", err)
		}
	}
	e.CreatedAt = ts.UTC()
	return e, nil
}
