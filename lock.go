package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/go-redis/redis/v8"
)

var (

	//go:embed script/lua/unlock.lua
	luaUnlock string

	ErrLockFailed  = errors.New("lock failed")
	ErrLockNotHold = errors.New("not hold current lock")
)

// Client is a client held redis client
type Client struct {
	client redis.Cmdable
}

// Lock is a lock created after client locks successfully
type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
}

// NewClient creates a new client
func NewClient(c redis.Cmdable) *Client {
	return &Client{
		client: c,
	}
}

// newLock creates a new lock
func newLock(c redis.Cmdable, key, value string, expiration time.Duration) *Lock {
	return &Lock{
		client:     c,
		key:        key,
		value:      value,
		expiration: expiration,
	}
}

// TryLock trys to hold a lock
func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	value := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		// Connect redis failed
		return nil, err
	}
	if !ok {
		// Lock held by other client
		return nil, ErrLockFailed
	}

	return newLock(c.client, key, value, expiration), nil
}

// UnLock unlocks
func (l *Lock) UnLock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value, l.expiration).Result()

	if err == redis.Nil {
		// Not current lock
		return ErrLockNotHold
	}
	if err != nil {
		// Connect redis failed
		return err
	}

	if res != 1 {
		// Not current lock
		return ErrLockNotHold
	}

	return nil
}
