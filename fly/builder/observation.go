package builder

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
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
	// Creates a txHashDedup to discard txHash from observations that were processed previously
	txHashDedup, err := NewDeduplicator("observations-dedup", config.ObservationsDedup, logger)
	if err != nil {
		return nil, err
	}
	cacheTxHash, err := NewCache[string]("observations-tx-hash", config.ObservationsTxHash.NumKeys, config.ObservationsTxHash.MaxCostsInMB)
	if err != nil {
		return nil, err
	}

	var txHashStores []txhash.TxHashStore
	expiration := time.Duration(config.ObservationsTxHash.ExpirationInSeconds) * time.Second
	txHashStores = append(txHashStores, txhash.NewCacheTxHash(cacheTxHash, expiration, logger))
	txHashStores = append(txHashStores, txhash.NewMongoTxHash(db, logger))
	txHashStore := txhash.NewComposite(txHashStores, metrics, logger)
	dedupTxHashStore := txhash.NewDedupTxHashStore(txHashStore, txHashDedup, logger)
	return dedupTxHashStore, nil
}
