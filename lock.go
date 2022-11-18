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
	//go:embed script/lua/refresh.lua
	luaRefresh string

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
	unlock     chan struct{}
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

// AutoRefresh always refreshes lock's expiration automatically
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	ch := make(chan struct{}, 1)
	defer close(ch)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-l.unlock:
			return nil
		case <-ch:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				// timeout, retry immediately
				ch <- struct{}{}
				continue
			}
			if err != nil {
				// network error, no way to resolve
				return err
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				// timeout, retry immediately
				ch <- struct{}{}
				continue
			}
			if err != nil {
				// network error, no way to resolve
				return err
			}
		}
	}
}

// Refresh refreshes lock's expiration
func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expiration.Milliseconds()).Int64()
	if err == redis.Nil {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	if res != int64(1) {
		return ErrLockNotHold
	}
	return nil
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
	defer func() {
		l.unlock <- struct{}{}
		close(l.unlock)
	}()

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
