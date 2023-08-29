package consumer

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc         queue.VAAConsumeFunc
	rpcProviderSettings *config.RpcProviderSettings
	logger              *zap.Logger
	repository          *Repository
	metrics             metrics.Metrics
	p2pNetwork          string
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	rpcProviderSettings *config.RpcProviderSettings,
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
	metrics metrics.Metrics,
	p2pNetwork string,
) *Consumer {

	c := Consumer{
		consumeFunc:         consumeFunc,
		rpcProviderSettings: rpcProviderSettings,
		logger:              logger,
		repository:          repository,
		metrics:             metrics,
		p2pNetwork:          p2pNetwork,
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go c.producerLoop(ctx)
}

func (c *Consumer) producerLoop(ctx context.Context) {

	ch := c.consumeFunc(ctx)

	for msg := range ch {
		c.logger.Debug("Received message", zap.String("vaaId", msg.Data().ID))
		c.process(ctx, msg)
	}
}

func (c *Consumer) process(ctx context.Context, msg queue.ConsumerMessage) {

	defer msg.Done()

	event := msg.Data()

	// Do not process messages from PythNet
	if event.ChainID == sdk.ChainIDPythNet {
		c.logger.Debug("Skipping expired PythNet message",
			zap.String("vaaId", event.ID),
			zap.String("trackId", event.TrackID))
		return
	}

	c.metrics.IncVaaUnfiltered(uint16(event.ChainID))

	// Process the VAA
	p := ProcessSourceTxParams{
		Timestamp: event.Timestamp,
		VaaId:     event.ID,
		ChainId:   event.ChainID,
		Emitter:   event.EmitterAddress,
		Sequence:  event.Sequence,
		TxHash:    event.TxHash,
		Overwrite: false, // avoid processing the same transaction twice
	}
	err := ProcessSourceTx(ctx, c.logger, c.rpcProviderSettings, c.repository, &p, c.p2pNetwork)

	// Log a message informing the processing status
	if errors.Is(err, chains.ErrChainNotSupported) {
		c.logger.Info("Skipping VAA - chain not supported",
			zap.String("vaaId", event.ID),
			zap.String("trackId", event.TrackID))
	} else if errors.Is(err, ErrAlreadyProcessed) {
		c.logger.Warn("Message already processed - skipping",
			zap.String("vaaId", event.ID),
			zap.String("trackId", event.TrackID))
	} else if errors.Is(err, ErrVaaWithoutTxHash) {
		c.logger.Error("Skipping VAA without txHash",
			zap.String("vaaId", event.ID),
			zap.String("trackId", event.TrackID))
	} else if err != nil {
		c.logger.Error("Failed to process originTx",
			zap.String("vaaId", event.ID),
			zap.Error(err),
			zap.String("trackId", event.TrackID))
	} else {
		c.logger.Info("Transaction processed successfully",
			zap.String("id", event.ID),
			zap.String("trackId", event.TrackID))
		c.metrics.IncOriginTxInserted(uint16(event.ChainID))
	}
}
