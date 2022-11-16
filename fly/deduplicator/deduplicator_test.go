package deduplicator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func newCache() cache.CacheInterface[bool] {
	c := gocache.New(5*time.Minute, 10*time.Minute)
	store := store.NewGoCache(c)
	return cache.New[bool](store)
}

func TestDeduplicator_Apply(t *testing.T) {
	ctx := context.TODO()
	c := newCache()
	logger := zaptest.NewLogger(t)
	d := New(c, logger)

	t.Run("single call", func(t *testing.T) {
		numberCalls := 0
		c.Clear(ctx)
		fnc := func() error {
			numberCalls++
			return nil
		}
		err := d.Apply(ctx, "key-1", fnc)
		assert.Nil(t, err)
		assert.Equal(t, 1, numberCalls)
	})

	t.Run("repeated key", func(t *testing.T) {
		numberCalls := 0
		c.Clear(ctx)
		fnc := func() error {
			numberCalls++
			return nil
		}
		err := d.Apply(ctx, "key-1", fnc)
		assert.Nil(t, err)
		err = d.Apply(ctx, "key-1", fnc)
		assert.Nil(t, err)
		err = d.Apply(ctx, "key-2", fnc)
		assert.Nil(t, err)
		err = d.Apply(ctx, "key-2", fnc)
		assert.Nil(t, err)
		assert.Equal(t, 2, numberCalls)
	})
}

func TestDeduplicator_Apply_Error(t *testing.T) {
	ctx := context.TODO()
	c := newCache()
	logger := zaptest.NewLogger(t)
	d := New(c, logger)

	t.Run("single call", func(t *testing.T) {
		numberCalls := 0
		c.Clear(ctx)
		fnc := func() error {
			numberCalls++
			return fmt.Errorf("failed")
		}
		err := d.Apply(ctx, "key-1", fnc)
		assert.NotNil(t, err)
		assert.Equal(t, 1, numberCalls)
	})

	t.Run("repeated key", func(t *testing.T) {
		numberCalls := 0
		c.Clear(ctx)
		fnc := func() error {
			numberCalls++
			return fmt.Errorf("failed")
		}
		err := d.Apply(ctx, "key-1", fnc)
		assert.NotNil(t, err)
		err = d.Apply(ctx, "key-1", fnc)
		assert.NotNil(t, err)
		err = d.Apply(ctx, "key-2", fnc)
		assert.NotNil(t, err)
		err = d.Apply(ctx, "key-2", fnc)
		assert.NotNil(t, err)
		assert.Equal(t, 4, numberCalls)
	})
}
