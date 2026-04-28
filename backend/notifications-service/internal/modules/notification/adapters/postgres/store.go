package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/domain"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) InsertNotification(ctx context.Context, in domain.SendInput) (*domain.Notification, error) {
	meta, _ := json.Marshal(in.Meta)
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO notifications.notifications_notifications(user_id,type,title,body,deep_link,meta)
         VALUES($1,$2,$3,$4,$5,$6)
         RETURNING id,user_id,type,title,body,deep_link,meta,read_at,created_at`,
		in.UserID, in.Type, in.Title, in.Body, in.DeepLink, meta)
	return scanNotif(row)
}

func (s *Store) ListNotifications(ctx context.Context, f domain.ListFilter) (*domain.NotificationsPage, error) {
	args := []any{f.UserID}
	q := `SELECT id,user_id,type,title,body,deep_link,meta,read_at,created_at FROM notifications.notifications_notifications WHERE user_id=$1`
	idx := 2

	if f.UnreadOnly {
		q += " AND read_at IS NULL"
	}
	if f.Cursor != "" {
		ts, id, err := decodeCursor(f.Cursor)
		if err == nil {
			q += fmt.Sprintf(" AND (created_at,id) < ($%d,$%d)", idx, idx+1)
			args = append(args, ts, id)
			idx += 2
		}
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", idx)
	args = append(args, f.Limit+1)

	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Notification
	for rows.Next() {
		n, err := scanNotif(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	page := &domain.NotificationsPage{}
	if len(items) > f.Limit {
		last := items[f.Limit-1]
		page.NextCursor = encodeCursor(last.CreatedAt, last.ID)
		items = items[:f.Limit]
	}
	page.Items = items
	return page, nil
}

func (s *Store) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE notifications.notifications_notifications SET read_at=now() WHERE id=$1 AND user_id=$2 AND read_at IS NULL`,
		id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE notifications.notifications_notifications SET read_at=now() WHERE user_id=$1 AND read_at IS NULL`,
		userID)
	return err
}

func (s *Store) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := s.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications.notifications_notifications WHERE user_id=$1 AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

func (s *Store) UpsertDevice(ctx context.Context, d domain.Device) (*domain.Device, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO notifications.notifications_devices(user_id,platform,push_token)
         VALUES($1,$2,$3)
         ON CONFLICT(push_token) DO UPDATE SET user_id=$1,last_seen_at=now()
         RETURNING id,user_id,platform,push_token,last_seen_at,created_at`,
		d.UserID, d.Platform, d.PushToken)
	var dev domain.Device
	if err := row.Scan(&dev.ID, &dev.UserID, &dev.Platform, &dev.PushToken, &dev.LastSeenAt, &dev.CreatedAt); err != nil {
		return nil, err
	}
	return &dev, nil
}

func (s *Store) DeleteDevice(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM notifications.notifications_devices WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.Preferences, error) {
	row := s.Pool.QueryRow(ctx,
		`SELECT user_id,push_trip_reminders,push_moderation_result,push_sync_status,push_marketing,
                email_welcome,email_moderation_result,email_marketing,updated_at
         FROM notifications.notifications_preferences WHERE user_id=$1`, userID)
	var p domain.Preferences
	if err := row.Scan(&p.UserID, &p.PushTripReminders, &p.PushModerationResult, &p.PushSyncStatus,
		&p.PushMarketing, &p.EmailWelcome, &p.EmailModerationResult, &p.EmailMarketing, &p.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpsertPreferences(ctx context.Context, p domain.Preferences) (*domain.Preferences, error) {
	row := s.Pool.QueryRow(ctx,
		`INSERT INTO notifications.notifications_preferences(user_id,push_trip_reminders,push_moderation_result,push_sync_status,push_marketing,email_welcome,email_moderation_result,email_marketing)
         VALUES($1,$2,$3,$4,$5,$6,$7,$8)
         ON CONFLICT(user_id) DO UPDATE SET
           push_trip_reminders=EXCLUDED.push_trip_reminders,
           push_moderation_result=EXCLUDED.push_moderation_result,
           push_sync_status=EXCLUDED.push_sync_status,
           push_marketing=EXCLUDED.push_marketing,
           email_welcome=EXCLUDED.email_welcome,
           email_moderation_result=EXCLUDED.email_moderation_result,
           email_marketing=EXCLUDED.email_marketing,
           updated_at=now()
         RETURNING user_id,push_trip_reminders,push_moderation_result,push_sync_status,push_marketing,email_welcome,email_moderation_result,email_marketing,updated_at`,
		p.UserID, p.PushTripReminders, p.PushModerationResult, p.PushSyncStatus,
		p.PushMarketing, p.EmailWelcome, p.EmailModerationResult, p.EmailMarketing)
	var out domain.Preferences
	if err := row.Scan(&out.UserID, &out.PushTripReminders, &out.PushModerationResult, &out.PushSyncStatus,
		&out.PushMarketing, &out.EmailWelcome, &out.EmailModerationResult, &out.EmailMarketing, &out.UpdatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func scanNotif(row pgx.Row) (*domain.Notification, error) {
	var n domain.Notification
	var meta []byte
	if err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.DeepLink, &meta, &n.ReadAt, &n.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	n.Meta = meta
	return &n, nil
}

func encodeCursor(t time.Time, id uuid.UUID) string {
	s := fmt.Sprintf("%s|%s", t.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func decodeCursor(c string) (time.Time, uuid.UUID, error) {
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("bad cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	return t, id, err
}
