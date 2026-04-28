package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

const slidingLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local n = redis.call('ZCARD', key)
if n < limit then
  redis.call('ZADD', key, now, member)
  redis.call('PEXPIRE', key, window)
  return {1, limit - n - 1, window}
end
local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
local reset = window
if #oldest == 2 then
  reset = window - (now - tonumber(oldest[2]))
  if reset < 0 then reset = 0 end
end
return {0, 0, reset}
`

// Client bundles Redis-backed auth helpers.
type Client struct {
	rdb *goredis.Client
}

func New(addr string) (*Client, error) {
	opt, err := goredis.ParseURL(addr)
	if err != nil {
		opt = &goredis.Options{Addr: addr}
	}
	rdb := goredis.NewClient(opt)
	return &Client{rdb: rdb}, nil
}

func (c *Client) RDB() *goredis.Client {
	return c.rdb
}

func (c *Client) Allow(ctx context.Context, key string, window time.Duration, limit int) (allowed bool, remaining int, reset time.Time, err error) {
	now := time.Now().UnixMilli()
	ms := int(window / time.Millisecond)
	if ms <= 0 {
		ms = 1000
	}
	member := uuid.NewString()
	res, err := c.rdb.Eval(ctx, slidingLua, []string{key}, now, ms, limit, member).Slice()
	if err != nil {
		return false, 0, time.Time{}, err
	}
	if len(res) < 3 {
		return false, 0, time.Time{}, errors.New("unexpected lua result")
	}
	ok, _ := res[0].(int64)
	rem, _ := res[1].(int64)
	resetMs, _ := res[2].(int64)
	allowed = ok == 1
	remaining = int(rem)
	reset = time.UnixMilli(now + resetMs)
	return allowed, remaining, reset, nil
}

func lockKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:lock:user:%s", userID)
}

func failKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:fail:user:%s", userID)
}

func (c *Client) IsLocked(ctx context.Context, userID uuid.UUID) (bool, time.Duration, error) {
	n, err := c.rdb.Exists(ctx, lockKey(userID)).Result()
	if err != nil {
		return false, 0, err
	}
	if n == 0 {
		return false, 0, nil
	}
	ttl, err := c.rdb.TTL(ctx, lockKey(userID)).Result()
	if err != nil {
		return true, 0, err
	}
	return true, ttl, nil
}

const maxFails = 5
const lockTTL = 15 * time.Minute
const failWindow = 15 * time.Minute

func (c *Client) RecordFailure(ctx context.Context, userID uuid.UUID) (locked bool, ttl time.Duration, err error) {
	k := failKey(userID)
	n, err := c.rdb.Incr(ctx, k).Result()
	if err != nil {
		return false, 0, err
	}
	if n == 1 {
		_ = c.rdb.Expire(ctx, k, failWindow).Err()
	}
	if n >= maxFails {
		if err := c.rdb.Set(ctx, lockKey(userID), "1", lockTTL).Err(); err != nil {
			return false, 0, err
		}
		return true, lockTTL, nil
	}
	return false, 0, nil
}

func (c *Client) Clear(ctx context.Context, userID uuid.UUID) error {
	pipe := c.rdb.Pipeline()
	pipe.Del(ctx, failKey(userID))
	pipe.Del(ctx, lockKey(userID))
	_, err := pipe.Exec(ctx)
	return err
}

func graceKey(hash string) string {
	return fmt.Sprintf("auth:refresh_grace:%s", hash)
}

func (c *Client) SetRefreshGrace(ctx context.Context, oldRefreshHash string, pair ports.TokenPair, ttl time.Duration) error {
	b, err := json.Marshal(pair)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, graceKey(oldRefreshHash), b, ttl).Err()
}

func (c *Client) GetRefreshGrace(ctx context.Context, oldRefreshHash string) (*ports.TokenPair, error) {
	s, err := c.rdb.Get(ctx, graceKey(oldRefreshHash)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p ports.TokenPair
	if err := json.Unmarshal(s, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func blKey(jti string) string {
	return fmt.Sprintf("auth:bl:jti:%s", jti)
}

func (c *Client) Add(ctx context.Context, jti string, ttl time.Duration) error {
	return c.rdb.Set(ctx, blKey(jti), "1", ttl).Err()
}

func (c *Client) Has(ctx context.Context, jti string) (bool, error) {
	n, err := c.rdb.Exists(ctx, blKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func resendKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:resend:verify:%s", userID)
}

func profileReservationKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:profile_reservation:%s", userID)
}

// SetProfileReservation stores pending Profile reservation until email is verified.
func (c *Client) SetProfileReservation(ctx context.Context, userID, reservationID uuid.UUID, ttl time.Duration) error {
	return c.rdb.Set(ctx, profileReservationKey(userID), reservationID.String(), ttl).Err()
}

// TakeProfileReservation returns and deletes the reservation id if present.
func (c *Client) TakeProfileReservation(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error) {
	key := profileReservationKey(userID)
	s, err := c.rdb.GetDel(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func (c *Client) ResendCooldown(ctx context.Context, userID uuid.UUID, cooldown time.Duration) (ok bool, retryAfter time.Duration, err error) {
	okSet, err := c.rdb.SetNX(ctx, resendKey(userID), "1", cooldown).Result()
	if err != nil {
		return false, 0, err
	}
	if okSet {
		return true, 0, nil
	}
	ttl, err := c.rdb.TTL(ctx, resendKey(userID)).Result()
	if err != nil {
		return false, 0, err
	}
	return false, ttl, nil
}
