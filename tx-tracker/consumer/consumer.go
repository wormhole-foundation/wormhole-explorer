package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc         queue.ConsumeFunc
	rpcProviderSettings *config.RpcProviderSettings
	logger              *zap.Logger
	repository          *Repository
	metrics             metrics.Metrics
	p2pNetwork          string
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.ConsumeFunc,
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
		c.logger.Debug("Skipping expired PythNet message", zap.String("trackId", event.TrackID), zap.String("vaaId", event.ID))
		return
	}

	if event.ChainID == sdk.ChainIDNear {
		c.logger.Warn("Skipping vaa from near", zap.String("trackId", event.TrackID), zap.String("vaaId", event.ID))
		return
	}

	start := time.Now()

	c.metrics.IncVaaUnfiltered(uint16(event.ChainID))

	// Process the VAA
	p := ProcessSourceTxParams{
		TrackID:   event.TrackID,
		Timestamp: event.Timestamp,
		VaaId:     event.ID,
		ChainId:   event.ChainID,
		Emitter:   event.EmitterAddress,
		Sequence:  event.Sequence,
		TxHash:    event.TxHash,
		Metrics:   c.metrics,
		Overwrite: false, // avoid processing the same transaction twice
	}
	_, err := ProcessSourceTx(ctx, c.logger, c.rpcProviderSettings, c.repository, &p, c.p2pNetwork)

	// add vaa processing duration metrics
	c.metrics.AddVaaProcessedDuration(uint16(event.ChainID), time.Since(start).Seconds())

	elapsedLog := zap.Uint64("elapsedTime", uint64(time.Since(start).Milliseconds()))
	// Log a message informing the processing status
	if errors.Is(err, chains.ErrChainNotSupported) {
		c.logger.Info("Skipping VAA - chain not supported",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			elapsedLog,
		)
	} else if errors.Is(err, ErrAlreadyProcessed) {
		c.logger.Warn("Message already processed - skipping",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			elapsedLog,
		)
	} else if err != nil {
		c.logger.Error("Failed to process originTx",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			zap.Error(err),
			elapsedLog,
		)
	} else {
		c.logger.Info("Transaction processed successfully",
			zap.String("trackId", event.TrackID),
			zap.String("id", event.ID),
			elapsedLog,
		)
		c.metrics.IncOriginTxInserted(uint16(event.ChainID))
	}
}
