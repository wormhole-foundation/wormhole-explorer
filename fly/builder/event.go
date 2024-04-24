package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/event"
	"go.uber.org/zap"
)

func NewEventDispatcher(ctx context.Context, config *config.Configuration, logger *zap.Logger) event.EventDispatcher {
	if config.IsLocal {
		return event.NewNoopEventDispatcher()
	}

	awsConfig, err := NewAwsConfig(ctx, config)
	if err != nil {
		logger.Fatal("could not create aws config", zap.Error(err))
	}

	ed, err := event.NewSnsEventDispatcher(awsConfig, config.Aws.EventsSnsUrl)
	if err != nil {
		logger.Fatal("could not create sns event dispatcher", zap.Error(err))
	}
	return ed
}
