package consumer

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc  queue.ConsumeFunc
	guardianPool *pool.Pool
	logger       *zap.Logger
	repository   *Repository
	metrics      metrics.Metrics
	p2pNetwork   string
	workersSize  int
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.ConsumeFunc,
	guardianPool *pool.Pool,
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
	metrics metrics.Metrics,
	p2pNetwork string,
	workersSize int,
) *Consumer {

	c := Consumer{
		consumeFunc:  consumeFunc,
		guardianPool: guardianPool,
		logger:       logger,
		repository:   repository,
		metrics:      metrics,
		p2pNetwork:   p2pNetwork,
		workersSize:  workersSize,
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

func (c *Consumer) producerLoop(ctx context.Context, ch <-chan queue.ConsumerMessage) {

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			fmt.Print(msg.Data()) //TODO: remove this line
			//TODO
		}
	}
}
