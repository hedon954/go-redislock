package go_redislock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hedon954/go-redislock/mocks"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestClient_TryLock(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name string

		mock func() redis.Cmdable

		// args
		key        string
		expiration time.Duration
		// expect res
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "locked",
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				rdb.EXPECT().
					SetNX(gomock.Any(), "locked-key", gomock.Any(), time.Minute).
					Return(res)

				return rdb
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
			name: "network error",
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, errors.New("network err"))
				rdb.EXPECT().
					SetNX(gomock.Any(), "network-key", gomock.Any(), time.Minute).
					Return(res)

				return rdb
			},
			key:        "network-key",
			expiration: time.Minute,
			wantErr:    errors.New("network err"),
		},
		{
			name: "failed to lock",
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, ErrLockFailed)
				rdb.EXPECT().
					SetNX(gomock.Any(), "fail-key", gomock.Any(), time.Minute).
					Return(res)

				return rdb
			},
			key:        "fail-key",
			expiration: time.Minute,
			wantErr:    ErrLockFailed,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rdb := tc.mock()
			c := NewClient(rdb)
			l, err := c.TryLock(context.Background(), tc.key, tc.expiration)
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
