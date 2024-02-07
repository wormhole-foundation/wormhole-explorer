package builder

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly/txhash"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Creates two callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
// callback to obtain queue messages from a queue
// callback to publish vaa non pyth messages to a sink
func NewObservationConsumePublish(ctx context.Context, config *config.Configuration, logger *zap.Logger) (health.Check, processor.ObservationQueueConsumeFunc, processor.ObservationPushFunc) {
	if config.IsLocal {
		obsQueue := queue.NewObservationInMemory()
		return health.Noop(), obsQueue.Consume, obsQueue.Publish
	}

	awsConfig, err := NewAwsConfig(ctx, config)
	if err != nil {
		logger.Fatal("could not create aws config", zap.Error(err))
	}

	sqsProducer, err := NewSQSProducer(awsConfig, config.Aws.ObservationsSqsUrl)
	if err != nil {
		logger.Fatal("could not create sqs producer", zap.Error(err))
	}

	sqsConsumer, err := NewSQSConsumer(config.Aws.ObservationsSqsUrl, ctx, config)
	if err != nil {
		logger.Fatal("could not create sqs consumer", zap.Error(err))
	}

	observationQueue := queue.NewObservationSqs(sqsProducer, sqsConsumer, logger)
	return health.SQS(awsConfig, config.Aws.ObservationsSqsUrl), observationQueue.Consume, observationQueue.Publish
}

func NewTxHashStore(ctx context.Context, config *config.Configuration, metrics metrics.Metrics, db *mongo.Database, logger *zap.Logger) (txhash.TxHashStore, error) {
	// Creates a deduplicator to discard VAA messages that were processed previously
	deduplicatorCache, err := NewCache[bool]()
	if err != nil {
		return nil, err
	}
	deduplicator := deduplicator.New(deduplicatorCache, logger)
	cacheTxHash, err := NewCache[txhash.TxHash]()
	if err != nil {
		return nil, err
	}

	var txHashStores []txhash.TxHashStore
	txHashStores = append(txHashStores, txhash.NewCacheTxHash(cacheTxHash, 30*time.Minute, logger))
	if !config.IsLocal {
		redisClient := NewRedisClient(config)
		txHashStores = append(txHashStores, txhash.NewRedisTxHash(redisClient, config.Redis.RedisPrefix, 30*time.Minute, logger))
	}
	txHashStores = append(txHashStores, txhash.NewMongoTxHash(db, logger))
	txHashStore := txhash.NewComposite(txHashStores, metrics, logger)
	dedupTxHashStore := txhash.NewDedupTxHashStore(txHashStore, deduplicator, logger)
	return dedupTxHashStore, nil
}
