package txhash

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/fly/domain"
	"go.uber.org/zap"
)

const txHashByVaaIDKey = "tx-hash-by-vaa-id"

type redisTxHash struct {
	client     *redis.Client
	prefix     string
	expiration time.Duration
	logger     *zap.Logger
}

func NewRedisTxHash(client *redis.Client,
	prefix string,
	expiration time.Duration,
	logger *zap.Logger) *redisTxHash {
	return &redisTxHash{
		client:     client,
		prefix:     prefix,
		expiration: expiration,
		logger:     logger,
	}
}

func (t *redisTxHash) Set(ctx context.Context, uniqueVaaID string, txHash TxHash) error {
	body, err := json.Marshal(txHash)
	if err != nil {
		return err
	}

	key := t.createKey(uniqueVaaID)
	if res := t.client.Set(ctx, key, string(body), t.expiration); res.Err() != nil {
		t.logger.Warn("Error setting tx hash in redis", zap.Error(res.Err()), zap.String("vaaId", uniqueVaaID))
		return res.Err()
	}

	return nil
}

func (r *redisTxHash) SetObservation(ctx context.Context, o *gossipv1.SignedObservation) error {
	txHash, err := CreateTxHash(r.logger, o)
	if err != nil {
		r.logger.Error("Error creating txHash", zap.Error(err))
		return err
	}
	uniqueVaaID := domain.CreateUniqueVaaIDByObservation(o)
	return r.Set(ctx, uniqueVaaID, *txHash)
}

func (r *redisTxHash) Get(ctx context.Context, uniqueVaaID string) (*TxHash, error) {
	key := r.createKey(uniqueVaaID)
	res := r.client.Get(ctx, key)
	if res.Err() == nil {
		var txHash TxHash
		err := json.Unmarshal([]byte(res.Val()), &txHash)
		if err != nil {
			return nil, err
		}
		return &txHash, nil
	}
	if res.Err() == redis.Nil {
		return nil, ErrTxHashNotFound
	}
	return nil, res.Err()
}

func (r *redisTxHash) createKey(uniqueVaaID string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, txHashByVaaIDKey, uniqueVaaID)
}

func (r *redisTxHash) GetName() string {
	return "redis"
}
