package cache

import (
	"context"
	"time"
)

// DummyCacheClient dummy cache client.
type DummyCacheClient struct {
}

// NewDummyCacheClient create a new instance of DummyCacheClient
func NewDummyCacheClient() *DummyCacheClient {
	return &DummyCacheClient{}
}

// Get get method is a dummy method that always does not find the cache.
// Use this Get function when run development enviroment
func (d *DummyCacheClient) Get(ctx context.Context, key string) (string, error) {
	return "", ErrNotFound
}

// Set set method is a dummy method that always does not set the cache.
func (d *DummyCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

// Close dummy cache client.
func (d *DummyCacheClient) Close() error {
	return nil
}
