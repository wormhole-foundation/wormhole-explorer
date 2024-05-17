package vaa

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	processor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc queue.ConsumeFunc[queue.EventDuplicateVaa]
	processor   processor.ProcessorFunc
	logger      *zap.Logger
	metrics     metrics.Metrics
	p2pNetwork  string
	workersSize int
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.ConsumeFunc[queue.EventDuplicateVaa],
	processor processor.ProcessorFunc,
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

func (c *Consumer) producerLoop(ctx context.Context, ch <-chan queue.ConsumerMessage[queue.EventDuplicateVaa]) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			c.processEvent(ctx, msg)
		}
	}
}

func (c *Consumer) processEvent(ctx context.Context, msg queue.ConsumerMessage[queue.EventDuplicateVaa]) {
	event := msg.Data()

	// Check if the event is a duplicate VAA event.
	if event.Type != queue.DeduplicateVaaEventType {
		msg.Done()
		c.logger.Debug("event is not a duplicate VAA",
			zap.Any("event", event))
		return
	}

	vaaID := event.Data.VaaID
	chainID := sdk.ChainID(event.Data.ChainID)

	logger := c.logger.With(
		zap.String("trackId", event.TrackID),
		zap.String("vaaId", vaaID))

	if msg.IsExpired() {
		msg.Failed()
		logger.Debug("event is expired")
		c.metrics.IncDuplicatedVaaExpired(chainID)
		return
	}

	params := &processor.Params{
		TrackID: event.TrackID,
		VaaID:   vaaID,
		ChainID: chainID,
	}

	err := c.processor(ctx, params)
	if err != nil {
		msg.Failed()
		logger.Error("error processing event", zap.Error(err))
		c.metrics.IncDuplicatedVaaFailed(chainID)
		return
	}

	msg.Done()
	logger.Debug("event processed")
	c.metrics.IncDuplicatedVaaProcessed(chainID)
}
