package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

func testClient(t *testing.T) *Client {
	t.Helper()
	s := miniredis.RunT(t)
	c, err := New("redis://" + s.Addr() + "/0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.RDB().Close() })
	return c
}

func TestAllow_SlidingWindow(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	key := "test:allow:" + t.Name()
	window := time.Minute
	limit := 3

	for i := 0; i < limit; i++ {
		ok, rem, _, err := c.Allow(ctx, key, window, limit)
		require.NoError(t, err)
		require.True(t, ok, "call %d", i)
		require.GreaterOrEqual(t, rem, 0)
	}
	ok, _, _, err := c.Allow(ctx, key, window, limit)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRecordFailure_Lockout(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	uid := uuid.New()

	for i := 0; i < maxFails-1; i++ {
		locked, _, err := c.RecordFailure(ctx, uid)
		require.NoError(t, err)
		require.False(t, locked)
	}
	locked, ttl, err := c.RecordFailure(ctx, uid)
	require.NoError(t, err)
	require.True(t, locked)
	require.Greater(t, ttl, time.Duration(0))

	is, d, err := c.IsLocked(ctx, uid)
	require.NoError(t, err)
	require.True(t, is)
	require.GreaterOrEqual(t, d, time.Duration(0))

	require.NoError(t, c.Clear(ctx, uid))
	is, _, err = c.IsLocked(ctx, uid)
	require.NoError(t, err)
	require.False(t, is)
}

func TestRefreshGraceRoundTrip(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	hash := "abc123"
	pair := ports.TokenPair{
		AccessToken:  "at",
		RefreshToken: "rt",
		AccessExp:    time.Now().Add(time.Minute).UTC(),
	}
	require.NoError(t, c.SetRefreshGrace(ctx, hash, pair, time.Minute))

	got, err := c.GetRefreshGrace(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, pair.AccessToken, got.AccessToken)
	require.Equal(t, pair.RefreshToken, got.RefreshToken)

	got2, err := c.GetRefreshGrace(ctx, hash)
	require.NoError(t, err)
	require.NotNil(t, got2)
	require.Equal(t, got.AccessToken, got2.AccessToken)
}

func TestAccessBlacklist(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	jti := "jti-test-1"
	require.NoError(t, c.Add(ctx, jti, time.Minute))
	ok, err := c.Has(ctx, jti)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestProfileReservation(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	uid := uuid.New()
	rid := uuid.New()
	require.NoError(t, c.SetProfileReservation(ctx, uid, rid, time.Minute))
	out, ok, err := c.TakeProfileReservation(ctx, uid)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, rid, out)
	_, ok2, err := c.TakeProfileReservation(ctx, uid)
	require.NoError(t, err)
	require.False(t, ok2)
}

func TestResendCooldown(t *testing.T) {
	ctx := context.Background()
	c := testClient(t)
	uid := uuid.New()
	cd := time.Minute

	ok, ra, err := c.ResendCooldown(ctx, uid, cd)
	require.NoError(t, err)
	require.True(t, ok)
	require.Zero(t, ra)

	ok2, ra2, err := c.ResendCooldown(ctx, uid, cd)
	require.NoError(t, err)
	require.False(t, ok2)
	require.Greater(t, ra2, time.Duration(0))
}
