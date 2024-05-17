package governor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	govprocessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/governor"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc queue.ConsumeFunc[queue.EventGovernorStatus]
	processor   govprocessor.ProcessorFunc
	logger      *zap.Logger
	metrics     metrics.Metrics
	p2pNetwork  string
	workersSize int
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.ConsumeFunc[queue.EventGovernorStatus],
	processor govprocessor.ProcessorFunc,
	logger *zap.Logger,
	metrics metrics.Metrics,
	p2pNetwork string,
	workersSize int,
) *Consumer {

	c := Consumer{
		consumeFunc: consumeFunc,
		processor:   processor,
		logger:      logger,
		metrics:     metrics,
		p2pNetwork:  p2pNetwork,
		workersSize: workersSize,
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

func (c *Consumer) producerLoop(ctx context.Context, ch <-chan queue.ConsumerMessage[queue.EventGovernorStatus]) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			c.processEvent(ctx, msg)
		}
	}
}

func (c *Consumer) processEvent(ctx context.Context, msg queue.ConsumerMessage[queue.EventGovernorStatus]) {
	event := msg.Data()

	// Check if the event is a governor status event.
	if event.Type != queue.GovernorStatusEventType {
		msg.Done()
		c.logger.Debug("event is not a governor status",
			zap.Any("event", event))
		return
	}

	params := &govprocessor.Params{
		TrackID:         event.TrackID,
		NodeGovernorVaa: domain.ConvertEventToGovernorVaa(&event),
	}

	err := c.processor(ctx, params)
	if err != nil {
		msg.Failed()
		c.logger.Error("failed to process governor-status event",
			zap.Error(err),
			zap.Any("event", event))
		// TODO: add metrics failed to process governor-status event.
		return
	}

	msg.Done()
	c.logger.Debug("governor-status event processed")
	// TODO: add metrics governor-status event processed.
}
