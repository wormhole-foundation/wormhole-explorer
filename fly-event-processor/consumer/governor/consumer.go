package governor

import (
	"context"

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

	// Check if a governor status message contains a new governor vaa.

	// Check if a governor status messages remove a vaa from governor.

	// err := c.processor(ctx, msg.Message.Params)
	// if err != nil {
	// 	c.logger.Error("Error processing event", zap.Error(err))
	// 	c.metrics.IncrementGovernorError()
	// 	return
	// }
	// c.metrics.IncrementGovernorProcessed()
}

type GovernorStatus struct {
	ChainID        uint32
	EmitterAddress string
	Sequence       uint64
	GovernorTxHash string
	ReleaseTime    uint64
	Amount         uint64
}
