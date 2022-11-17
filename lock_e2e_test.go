//go:build e2e
// +build e2e

package go_redislock

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClientE2ESuite struct {
	suite.Suite
	rdb redis.Cmdable
}

func (s *ClientE2ESuite) SetupSuite() {
	s.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})

	// 确保测试的目标 Redis 已经成功启动了
	for s.rdb.Ping(context.Background()).Err() != nil {

	}
}

func TestClient_TryLock_e2e(t *testing.T) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})

	testcases := []struct {
		name string

		// 准备数据
		before func()
		// 校验 Redis 数据并且清理数据
		after func()

		// args
		key        string
		expiration time.Duration
		// expect res
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "locked",

			before: func() {},
			after: func() {
				res, err := rdb.Del(context.Background(), "locked-key").Result()
				require.NoError(t, err)
				require.Equal(t, int64(1), res)
			},

			key:        "locked-key",
			expiration: time.Minute,
			wantErr:    nil,
			wantLock: &Lock{
				key:        "locked-key",
				expiration: time.Minute,
			},
		},
		{
			name: "failed to lock",
			before: func() {
				res, err := rdb.Set(context.Background(), "failed-key", "123", time.Minute).Result()
				require.NoError(t, err)
				require.Equal(t, "OK", res)
			},
			after: func() {
				res, err := rdb.Get(context.Background(), "failed-key").Result()
				require.NoError(t, err)
				require.Equal(t, "123", res)
				result, err := rdb.Del(context.Background(), "failed-key").Result()
				require.NoError(t, err)
				require.Equal(t, int64(1), result)
			},
			key:        "failed-key",
			expiration: time.Minute,
			wantErr:    ErrLockFailed,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			c := NewClient(rdb)
			l, err := c.TryLock(context.Background(), tc.key, tc.expiration)
			tc.after()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			empty := assert.NotEmpty(t, l)
			if empty {
				return
			}
			assert.Equal(t, l.key, tc.key)
		})
	}

}
