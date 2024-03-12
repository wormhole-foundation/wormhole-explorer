package txhash

import (
	"context"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"go.uber.org/zap"
)

type cacheTxHash struct {
	cache      cache.CacheInterface[string]
	expiration time.Duration
	logger     *zap.Logger
}

func NewCacheTxHash(cache cache.CacheInterface[string],
	expiration time.Duration,
	logger *zap.Logger) *cacheTxHash {
	return &cacheTxHash{
		cache:      cache,
		expiration: expiration,
		logger:     logger,
	}
}

func (t *cacheTxHash) Set(ctx context.Context, vaaID string, txHash TxHash) error {
	if err := t.cache.Set(ctx, vaaID, txHash.TxHash, store.WithCost(256), store.WithExpiration(t.expiration)); err != nil {
		t.logger.Error("Error setting tx hash in cache", zap.Error(err))
		return err
	}
	return nil
}

func (r *cacheTxHash) SetObservation(ctx context.Context, o *gossipv1.SignedObservation) error {
	txHash, err := CreateTxHash(r.logger, o)
	if err != nil {
		r.logger.Error("Error creating txHash", zap.Error(err))
		return err
	}
	return r.Set(ctx, o.MessageId, *txHash)
}

func (r *cacheTxHash) Get(ctx context.Context, vaaID string) (*string, error) {
	txHash, err := r.cache.Get(ctx, vaaID)
	if err == nil {
		return &txHash, nil
	}
	return nil, ErrTxHashNotFound
}

func (r *cacheTxHash) GetName() string {
	return "memory"
}
