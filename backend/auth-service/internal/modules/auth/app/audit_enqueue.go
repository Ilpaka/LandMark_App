package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/audit"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// enqueueAudit writes a masked audit entry to the transactional outbox (processed by OutboxPoller).
func (s *Service) enqueueAudit(ctx context.Context, tx pgx.Tx, e ports.AuditEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if e.Meta == nil {
		e.Meta = map[string]any{}
	}
	masked := audit.MaskEvent(e)
	pl, err := audit.EventToMap(masked)
	if err != nil {
		return err
	}
	return s.Store.InsertOutbox(ctx, tx, ports.OutboxMessage{
		AggregateType: "audit",
		EventType:     "audit.append",
		Payload:       pl,
	})
}
