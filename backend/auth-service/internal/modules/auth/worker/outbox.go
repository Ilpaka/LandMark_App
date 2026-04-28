package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/audit"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// OutboxPoller publishes unpublished outbox rows: audit.append → auth_audit_log, others logged (future: message bus).
func OutboxPoller(ctx context.Context, store ports.Store, log *slog.Logger, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
				rows, err := store.ListUnpublishedOutbox(ctx, tx, 50)
				if err != nil {
					return err
				}
				now := time.Now().UTC()
				for _, row := range rows {
					switch {
					case row.AggregateType == "audit" && row.EventType == "audit.append":
						ev, err := audit.MapToEvent(row.Payload)
						if err != nil {
							log.Error("audit outbox decode", "id", row.ID, "err", err)
							return err
						}
						if err := store.InsertAudit(ctx, tx, ev); err != nil {
							log.Error("audit insert", "id", row.ID, "err", err)
							return err
						}
					default:
						log.Info("outbox publish",
							"id", row.ID,
							"aggregate_type", row.AggregateType,
							"event_type", row.EventType,
						)
					}
					if err := store.MarkOutboxPublished(ctx, tx, row.ID, now); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
}
