package governor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	govConfigProcessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/governor_config"
	govStatusProcessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/governor_status"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc        queue.ConsumeFunc[queue.EventGovernor]
	govStatusProcessor govStatusProcessor.ProcessorFunc
	govConfigProcessor govConfigProcessor.ProcessorFunc
	logger             *zap.Logger
	metrics            metrics.Metrics
	p2pNetwork         string
	workersSize        int
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.ConsumeFunc[queue.EventGovernor],
	govStatusProcessor govStatusProcessor.ProcessorFunc,
	govConfigProcessor govConfigProcessor.ProcessorFunc,
	logger *zap.Logger,
	metrics metrics.Metrics,
	p2pNetwork string,
	workersSize int,
) *Consumer {

	c := Consumer{
		consumeFunc:        consumeFunc,
		govStatusProcessor: govStatusProcessor,
		govConfigProcessor: govConfigProcessor,
		logger:             logger,
		metrics:            metrics,
		p2pNetwork:         p2pNetwork,
		workersSize:        workersSize,
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	ch := c.consumeFunc(ctx)
	for i := 0; i < c.workersSize; i++ {
		go c.producerLoop(ctx, ch)
	}
}

func (c *Consumer) producerLoop(ctx context.Context, ch <-chan queue.ConsumerMessage[queue.EventGovernor]) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			event := msg.Data()
			switch event.Type {
			case queue.GovernorStatusEventType:
				c.processGovStatusEvent(ctx, msg)
			case queue.GovernorConfigEventType:
				c.processGovConfigEvent(ctx, msg)
			default:
				c.logger.Debug("event is not a governor message",
					zap.Any("event", event))
				msg.Done()
			}
		}
	}
}

func (c *Consumer) processGovStatusEvent(ctx context.Context, msg queue.ConsumerMessage[queue.EventGovernor]) {
	// get governor status event
	event := msg.Data()
	govStatusEvent, ok := event.Data.(queue.GovernorStatus)
	if !ok {
		msg.Done()
		c.logger.Debug("event data is not a governor status",
			zap.Any("event", event))
		return
	}

	logger := c.logger.With(
		zap.String("trackId", event.TrackID),
		zap.String("type", event.Type),
		zap.String("node", govStatusEvent.NodeName))

	// check if event is expired
	if msg.IsExpired() {
		msg.Failed()
		logger.Debug("event is expired")
		c.metrics.IncGovernorStatusExpired(govStatusEvent.NodeName,
			govStatusEvent.NodeAddress)
		return
	}

	// process governor status event
	params := &govStatusProcessor.Params{
		TrackID: event.TrackID,
		//NodeGovernorVaa: domain.ConvertEventToGovernorVaa(&event),
		NodeGovernorVaa: domain.ConvertEventToGovernorVaa(&govStatusEvent),
	}
	err := c.govStatusProcessor(ctx, params)
	if err != nil {
		msg.Failed()
		logger.Error("failed to process governor-status event", zap.Error(err))
		c.metrics.IncGovernorStatusFailed(params.NodeGovernorVaa.Name, params.NodeGovernorVaa.Address)
		return
	}

	msg.Done()
	logger.Debug("governor-status event processed")
	c.metrics.IncGovernorStatusProcessed(params.NodeGovernorVaa.Name, params.NodeGovernorVaa.Address)
}

func (c *Consumer) processGovConfigEvent(ctx context.Context, msg queue.ConsumerMessage[queue.EventGovernor]) {
	// get governor config event
	event := msg.Data()
	govConfigEvent, ok := event.Data.(queue.GovernorConfig)
	if !ok {
		msg.Done()
		c.logger.Debug("event data is not a governor config",
			zap.Any("event", event))
		return
	}

	logger := c.logger.With(
		zap.String("trackId", event.TrackID),
		zap.String("type", event.Type),
		zap.String("node", govConfigEvent.NodeName))

	// check if event is expired
	if msg.IsExpired() {
		msg.Failed()
		logger.Debug("event is expired")
		c.metrics.IncGovernorConfigExpired(govConfigEvent.NodeName,
			govConfigEvent.NodeAddress)
		return
	}

	// process governor config event
	params := &govConfigProcessor.Params{
		TrackID:        event.TrackID,
		GovernorConfig: govConfigEvent,
	}

	err := c.govConfigProcessor(ctx, params)
	if err != nil {
		msg.Failed()
		logger.Error("failed to process governor-config event", zap.Error(err))
		c.metrics.IncGovernorConfigFailed(govConfigEvent.NodeName, govConfigEvent.NodeAddress)
		return
	}

	msg.Done()
	logger.Debug("governor-config event processed")
	c.metrics.IncGovernorConfigProcessed(govConfigEvent.NodeName, govConfigEvent.NodeAddress)
}
