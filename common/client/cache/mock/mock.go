package mock

import (
	"context"
	"github.com/test-go/testify/mock"
	"time"
)

// CacheMock exported type to provide mock for cache.Cache interface
type CacheMock struct {
	mock.Mock
}

func (c *CacheMock) Get(ctx context.Context, key string) (string, error) {
	args := c.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (c *CacheMock) Close() error {
	return nil
}

func (c *CacheMock) Set(ctx context.Context, key string, value interface{}, expirations time.Duration) error {
	args := c.Called(ctx, key, value, expirations)
	return args.Error(0)
}
