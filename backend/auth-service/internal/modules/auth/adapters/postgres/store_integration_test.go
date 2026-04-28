//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
	"github.com/ilpaka/landmark_app/backend/auth-service/tests/testutil"
)

func TestStore_InsertAccountAndGetByEmailNorm(t *testing.T) {
	deps := testutil.RequireAuthDocker(t)
	st := New(deps.Pool)
	ctx := context.Background()

	var insertedID uuid.UUID
	err := st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := st.InsertAccount(ctx, tx, "Ins@example.com", "ins@example.com", domain.StatusPendingVerification, domain.RoleUser)
		if err != nil {
			return err
		}
		insertedID = acc.ID
		got, err := st.GetAccountByEmailNorm(ctx, tx, "ins@example.com")
		if err != nil {
			return err
		}
		require.Equal(t, acc.ID, got.ID)
		return nil
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, insertedID)
}

func TestStore_OutboxInsertListMark(t *testing.T) {
	deps := testutil.RequireAuthDocker(t)
	st := New(deps.Pool)
	ctx := context.Background()

	err := st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := st.InsertOutbox(ctx, tx, ports.OutboxMessage{
			AggregateType: "user",
			EventType:     "user.test",
			Payload:       map[string]any{"k": "v"},
		}); err != nil {
			return err
		}
		rows, err := st.ListUnpublishedOutbox(ctx, tx, 10)
		if err != nil {
			return err
		}
		require.NotEmpty(t, rows)
		id := rows[len(rows)-1].ID
		return st.MarkOutboxPublished(ctx, tx, id, time.Now().UTC())
	})
	require.NoError(t, err)
}

func TestStore_InsertSessionAndGetByRefreshHash(t *testing.T) {
	deps := testutil.RequireAuthDocker(t)
	st := New(deps.Pool)
	ctx := context.Background()
	fam := uuid.New()
	sid := uuid.New()
	hash := "deadbeef" + uuid.NewString()

	err := st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := st.InsertAccount(ctx, tx, "sess@example.com", "sess@example.com", domain.StatusActive, domain.RoleUser)
		if err != nil {
			return err
		}
		acc, err := st.GetAccountByEmailNorm(ctx, tx, "sess@example.com")
		if err != nil {
			return err
		}
		return st.InsertSession(ctx, tx, ports.Session{
			ID:               sid,
			UserID:           acc.ID,
			RefreshTokenHash: hash,
			SessionFamilyID:  fam,
			IssuedAt:         time.Now().UTC(),
			ExpiresAt:        time.Now().Add(24 * time.Hour).UTC(),
		})
	})
	require.NoError(t, err)

	err = st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		sess, err := st.GetSessionByRefreshHashForUpdate(ctx, tx, hash)
		if err != nil {
			return err
		}
		require.Equal(t, sid, sess.ID)
		require.Equal(t, hash, sess.RefreshTokenHash)
		return nil
	})
	require.NoError(t, err)
}
