package cacheable_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func Test_GetOrLoad(t *testing.T) {

	const ttl = 10 * time.Minute

	tests := []struct {
		name                  string
		key                   string
		cacheResult           cacheable.CachedResult[int]
		cacheErr              error
		expectedCacheGetCalls int
		expectedCacheSetCalls int
		load                  func() (int, error)
		want                  interface{}
		wantErr               bool
		opts                  []cacheable.Opts
	}{
		{
			name: "Cache hit, valid entry",
			key:  "key1",
			load: func() (int, error) {
				return 0, nil
			},
			cacheResult: cacheable.CachedResult[int]{
				Result:    42,
				Timestamp: time.Now().Add(-ttl / 2),
			},
			expectedCacheGetCalls: 1,
			expectedCacheSetCalls: 0,
			want:                  42,
			wantErr:               false,
		},
		{
			name: "Cache miss, load success",
			key:  "key2",
			load: func() (int, error) {
				return 123, nil
			},
			cacheErr:              cache.ErrNotFound,
			expectedCacheGetCalls: 1,
			expectedCacheSetCalls: 1,
			want:                  123,
			wantErr:               false,
		},
		{
			name: "Cache error, load success",
			key:  "key3",
			load: func() (int, error) {
				return 789, nil
			},
			cacheErr:              errors.New("mocked_cache_error"),
			expectedCacheGetCalls: 1,
			expectedCacheSetCalls: 1,
			want:                  789,
			wantErr:               false,
		},
		{
			name: "Cache hit, expired entry, no auto-renew, load fails",
			key:  "key4",
			load: func() (int, error) {
				return 0, errors.New("load error")
			},
			cacheErr: nil,
			cacheResult: cacheable.CachedResult[int]{
				Result:    100,
				Timestamp: time.Now().Add(-ttl * 2),
			},
			expectedCacheGetCalls: 1,
			expectedCacheSetCalls: 0,
			want:                  100,
			wantErr:               false,
		},
		{
			name: "Cache hit, expired entry, auto-renew, load success",
			key:  "key5",
			load: func() (int, error) {
				return 300, nil
			},
			cacheErr: nil,
			cacheResult: cacheable.CachedResult[int]{
				Result:    100,
				Timestamp: time.Now().Add(-ttl * 2),
			},
			expectedCacheGetCalls: 1,
			expectedCacheSetCalls: 1,
			want:                  100,
			wantErr:               false,
			opts:                  []cacheable.Opts{cacheable.WithAutomaticRenew()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCache := new(mockCacheClient)
			bytes, _ := json.Marshal(tt.cacheResult)
			mockCache.On("Get", ctx, tt.key).Return(string(bytes), tt.cacheErr)
			mockCache.On("Set", mock.Anything, tt.key, mock.Anything, time.Duration(0)).Return(nil)

			result, err := cacheable.GetOrLoad[int](
				ctx,
				zaptest.NewLogger(t),
				mockCache,
				ttl,
				tt.key,
				metrics.NewNoOpMetrics(),
				tt.load,
				tt.opts...,
			)

			time.Sleep(100 * time.Millisecond)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
			mockCache.AssertNumberOfCalls(t, "Get", tt.expectedCacheGetCalls)
			mockCache.AssertNumberOfCalls(t, "Set", tt.expectedCacheSetCalls)
		})
	}
}

type mockCacheClient struct {
	mock.Mock
}

func (m *mockCacheClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockCacheClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockCacheClient) Set(ctx context.Context, key string, value interface{}, expirations time.Duration) error {
	args := m.Called(ctx, key, value, expirations)
	return args.Error(0)
}
