package builder

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/notifier"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly/producer"
	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Creates a callback to publish VAA messages to a redis pubsub
func NewVAARedisProducerFunc(cfg *config.Configuration, logger *zap.Logger) (producer.PushFunc, error) {
	if cfg.IsLocal {
		return func(context.Context, *producer.Notification) error {
			return nil
		}, nil
	}
	client := NewRedisClient(cfg)
	channel := fmt.Sprintf("%s:%s", cfg.Redis.RedisPrefix, cfg.Redis.RedisVaaChannel)
	logger.Info("using redis producer", zap.String("channel", channel))
	return producer.NewRedisProducer(client, channel).Push, nil
}

// Creates two callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
// callback to obtain queue messages from a queue
// callback to publish vaa non pyth messages to a sink
func NewVAAConsumePublish(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (health.Check, processor.VAAQueueConsumeFunc, processor.VAAPushFunc) {
	if cfg.IsLocal {
		vaaQueue := queue.NewVAAInMemory()
		return health.Noop(), vaaQueue.Consume, vaaQueue.Publish
	}

	awsConfig, err := NewAwsConfig(ctx, cfg)
	if err != nil {
		logger.Fatal("could not create aws config", zap.Error(err))
	}

	sqsProducer, err := NewSQSProducer(awsConfig, cfg.Aws.SqsUrl)
	if err != nil {
		logger.Fatal("could not create sqs producer", zap.Error(err))
	}

	sqsConsumer, err := NewSQSConsumer(cfg.Aws.SqsUrl, ctx, cfg)
	if err != nil {
		logger.Fatal("could not create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewVaaSqs(sqsProducer, sqsConsumer, logger)
	return health.SQS(awsConfig, cfg.Aws.SqsUrl), vaaQueue.Consume, vaaQueue.Publish
}

func NewVAANotifierFunc(cfg *config.Configuration, logger *zap.Logger) processor.VAANotifyFunc {
	if cfg.IsLocal {
		return func(context.Context, *vaa.VAA, []byte) error {
			return nil
		}
	}

	logger.Info("using redis notifier", zap.String("prefix", cfg.Redis.RedisPrefix))
	client := redis.NewClient(&redis.Options{Addr: cfg.Redis.RedisUri})

	return notifier.NewLastSequenceNotifier(client, cfg.Redis.RedisPrefix).Notify
}
