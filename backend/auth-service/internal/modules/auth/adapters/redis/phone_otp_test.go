package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

func TestPhoneOTP_CancelAndVerify_Success(t *testing.T) {
	t.Parallel()
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	c := &Client{rdb: rdb}
	ctx := context.Background()

	userID := uuid.New()
	salt, err := crypto.RandomSalt(16)
	require.NoError(t, err)
	plain := "123456"
	hash := crypto.HashOTP(salt, plain)

	id, err := c.CancelPendingAndCreate(ctx, userID, "+79991234567", domain.PhoneOTPPurposeLogin, salt, hash, 15*time.Minute)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	payload, err := c.VerifyAndConsume(ctx, id, domain.PhoneOTPPurposeLogin, plain)
	require.NoError(t, err)
	require.Equal(t, userID, payload.UserID)
	require.Equal(t, "+79991234567", payload.PhoneE164)

	_, err = c.VerifyAndConsume(ctx, id, domain.PhoneOTPPurposeLogin, plain)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPhoneOTP_WrongCode_IncrementsAttempts(t *testing.T) {
	t.Parallel()
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	c := &Client{rdb: rdb}
	ctx := context.Background()
	userID := uuid.New()
	salt, _ := crypto.RandomSalt(16)
	plain := "999999"
	hash := crypto.HashOTP(salt, "111111")
	id, err := c.CancelPendingAndCreate(ctx, userID, "+7000", domain.PhoneOTPPurposeLogin, salt, hash, 15*time.Minute)
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		_, err := c.VerifyAndConsume(ctx, id, domain.PhoneOTPPurposeLogin, plain)
		require.ErrorIs(t, err, domain.ErrInvalidOTP, "attempt %d", i)
	}
	_, err = c.VerifyAndConsume(ctx, id, domain.PhoneOTPPurposeLogin, plain)
	require.ErrorIs(t, err, domain.ErrTooManyAttempts)
}

func TestPhoneOTP_ReplacePending(t *testing.T) {
	t.Parallel()
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	c := &Client{rdb: rdb}
	ctx := context.Background()
	userID := uuid.New()
	s1, _ := crypto.RandomSalt(16)
	h1 := crypto.HashOTP(s1, "111111")
	id1, err := c.CancelPendingAndCreate(ctx, userID, "+7000", domain.PhoneOTPPurposeLogin, s1, h1, 15*time.Minute)
	require.NoError(t, err)
	s2, _ := crypto.RandomSalt(16)
	h2 := crypto.HashOTP(s2, "222222")
	id2, err := c.CancelPendingAndCreate(ctx, userID, "+7000", domain.PhoneOTPPurposeLogin, s2, h2, 15*time.Minute)
	require.NoError(t, err)
	require.NotEqual(t, id1, id2)
	_, err = c.VerifyAndConsume(ctx, id1, domain.PhoneOTPPurposeLogin, "111111")
	require.ErrorIs(t, err, domain.ErrNotFound)
	p, err := c.VerifyAndConsume(ctx, id2, domain.PhoneOTPPurposeLogin, "222222")
	require.NoError(t, err)
	require.Equal(t, userID, p.UserID)
}
