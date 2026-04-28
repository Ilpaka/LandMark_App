package redis

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

const (
	phoneOTPKeyVersion = "v1"
	phoneOTPChPrefix   = "auth:otp:" + phoneOTPKeyVersion + ":ch:"
	phoneOTPPenPrefix  = "auth:otp:" + phoneOTPKeyVersion + ":pen:"
	defaultMaxAttempts = 5
)

func challengeKey(id uuid.UUID) string {
	return phoneOTPChPrefix + id.String()
}

func pendingKey(userID uuid.UUID) string {
	return phoneOTPPenPrefix + userID.String()
}

// PhoneOTPCancelAndCreateScript cancels pending challenges and stores a new one.
const phoneOTPCancelAndCreateScript = `
local penKey = KEYS[1]
local chKey = KEYS[2]
local ttl = tonumber(ARGV[1])
local userId = ARGV[2]
local phone = ARGV[3]
local purpose = ARGV[4]
local saltB64 = ARGV[5]
local codeHash = ARGV[6]
local maxAttempts = ARGV[7]
local vid = ARGV[8]
local chPrefix = ARGV[9]

local members = redis.call('SMEMBERS', penKey)
for _, m in ipairs(members) do
  redis.call('DEL', chPrefix .. m)
end
redis.call('DEL', penKey)
redis.call('HSET', chKey,
  'user_id', userId,
  'phone_e164', phone,
  'purpose', purpose,
  'code_salt', saltB64,
  'code_hash', codeHash,
  'attempts', '0',
  'max_attempts', maxAttempts
)
redis.call('PEXPIRE', chKey, ttl)
redis.call('SADD', penKey, vid)
redis.call('PEXPIRE', penKey, ttl)
return 1
`

// PhoneOTPVerifyScript checks code hash, updates attempts, deletes on success.
const phoneOTPVerifyScript = `
local k = KEYS[1]
local expPurpose = ARGV[1]
local candidateHash = ARGV[2]
local vid = ARGV[3]
local penPrefix = ARGV[4]

if redis.call('EXISTS', k) == 0 then
  return {0}
end
local p = redis.call('HGET', k, 'purpose')
if p ~= expPurpose then
  return {1}
end
local att = tonumber(redis.call('HGET', k, 'attempts') or '0')
local maxa = tonumber(redis.call('HGET', k, 'max_attempts') or '5')
if att >= maxa then
  return {2}
end
local h = redis.call('HGET', k, 'code_hash')
if h ~= candidateHash then
  redis.call('HINCRBY', k, 'attempts', 1)
  return {3}
end
local uid = redis.call('HGET', k, 'user_id')
local ph = redis.call('HGET', k, 'phone_e164')
local penKey = penPrefix .. uid
redis.call('SREM', penKey, vid)
redis.call('DEL', k)
return {4, uid, ph}
`

// CancelPendingAndCreate implements [ports.PhoneOTPChallengeStore] for Redis.
func (c *Client) CancelPendingAndCreate(
	ctx context.Context,
	userID uuid.UUID,
	phoneE164, purpose string,
	salt []byte,
	codeHash string,
	ttl time.Duration,
) (uuid.UUID, error) {
	verificationID := uuid.New()
	vid := verificationID.String()
	ttlMs := ttl.Milliseconds()
	if ttlMs <= 0 {
		ttlMs = 1
	}
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	pen := pendingKey(userID)
	ch := challengeKey(verificationID)
	_, err := c.rdb.Eval(ctx, phoneOTPCancelAndCreateScript, []string{pen, ch},
		ttlMs,
		userID.String(),
		phoneE164,
		purpose,
		saltB64,
		codeHash,
		fmt.Sprint(defaultMaxAttempts),
		vid,
		phoneOTPChPrefix,
	).Int()
	if err != nil {
		return uuid.Nil, err
	}
	return verificationID, nil
}

// VerifyAndConsume implements [ports.PhoneOTPChallengeStore] for Redis.
func (c *Client) VerifyAndConsume(
	ctx context.Context,
	verificationID uuid.UUID,
	purpose, plainCode string,
) (*ports.PhoneOTPVerifiedPayload, error) {
	k := challengeKey(verificationID)
	raw, err := c.rdb.HGetAll(ctx, k).Result()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, domain.ErrNotFound
	}
	if raw["purpose"] != purpose {
		return nil, domain.ErrConflict
	}
	att := parseIntField(raw["attempts"], 0)
	maxa := parseIntField(raw["max_attempts"], defaultMaxAttempts)
	if maxa <= 0 {
		maxa = defaultMaxAttempts
	}
	if att >= maxa {
		return nil, domain.ErrTooManyAttempts
	}
	salt, derr := base64.StdEncoding.DecodeString(raw["code_salt"])
	if derr != nil {
		return nil, domain.ErrInvalidOTP
	}
	candidateHash := crypto.HashOTP(salt, plainCode)
	vid := verificationID.String()
	res, err := c.rdb.Eval(ctx, phoneOTPVerifyScript, []string{k},
		purpose,
		candidateHash,
		vid,
		phoneOTPPenPrefix,
	).Result()
	if err != nil {
		return nil, err
	}
	arr, ok := res.([]any)
	if !ok || len(arr) < 1 {
		return nil, errors.New("redis: unexpected verify result")
	}
	status := intFromRedisAny(arr[0])
	switch status {
	case 0:
		return nil, domain.ErrNotFound
	case 1:
		return nil, domain.ErrConflict
	case 2:
		return nil, domain.ErrTooManyAttempts
	case 3:
		return nil, domain.ErrInvalidOTP
	case 4:
		if len(arr) < 3 {
			return nil, errors.New("redis: verify success without payload")
		}
		uidStr, _ := arr[1].(string)
		phone, _ := arr[2].(string)
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			return nil, err
		}
		return &ports.PhoneOTPVerifiedPayload{UserID: uid, PhoneE164: phone}, nil
	default:
		return nil, errors.New("redis: unknown verify status")
	}
}

func parseIntField(s string, def int64) int64 {
	if s == "" {
		return def
	}
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int64(c-'0')
	}
	return n
}

func intFromRedisAny(v any) int {
	switch t := v.(type) {
	case int64:
		return int(t)
	case int:
		return t
	case float64:
		return int(t)
	default:
		return -1
	}
}
