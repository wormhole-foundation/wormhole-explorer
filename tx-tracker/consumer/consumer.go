package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc      queue.ConsumeFunc
	rpcpool          map[vaa.ChainID]*pool.Pool
	wormchainRpcPool map[vaa.ChainID]*pool.Pool
	logger           *zap.Logger
	repository       Repository
	metrics          metrics.Metrics
	p2pNetwork       string
	workersSize      int
	notionalCache    *notional.NotionalCache
}

// New creates a new vaa consumer.
func New(consumeFunc queue.ConsumeFunc,
	rpcPool map[sdk.ChainID]*pool.Pool,
	wormchainRpcPool map[sdk.ChainID]*pool.Pool,
	logger *zap.Logger,
	repository Repository,
	metrics metrics.Metrics,
	p2pNetwork string,
	workersSize int,
	notionalCache *notional.NotionalCache,
) *Consumer {

	c := Consumer{
		consumeFunc:      consumeFunc,
		rpcpool:          rpcPool,
		wormchainRpcPool: wormchainRpcPool,
		logger:           logger,
		repository:       repository,
		metrics:          metrics,
		p2pNetwork:       p2pNetwork,
		workersSize:      workersSize,
		notionalCache:    notionalCache,
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
			c.logger.Debug("Received message", zap.String("vaaId", msg.Data().ID), zap.String("trackId", msg.Data().TrackID))
			switch msg.Data().Type {
			case queue.SourceChainEvent:
				c.processSourceTx(ctx, msg)
			case queue.TargetChainEvent:
				c.processTargetTx(ctx, msg)
			default:
				c.logger.Error("Unknown message type", zap.String("trackId", msg.Data().TrackID), zap.Any("type", msg.Data().Type))
			}
		}
	}
}

func (c *Consumer) processSourceTx(ctx context.Context, msg queue.ConsumerMessage) {

	event := msg.Data()

	// Do not process messages from PythNet
	if event.ChainID == sdk.ChainIDPythNet {
		msg.Done()
		c.logger.Debug("Skipping pythNet message", zap.String("trackId", event.TrackID), zap.String("vaaId", event.ID))
		return
	}

	if event.ChainID == sdk.ChainIDNear {
		msg.Done()
		c.logger.Warn("Skipping vaa from near", zap.String("trackId", event.TrackID), zap.String("vaaId", event.ID))
		return
	}

	start := time.Now()

	c.metrics.IncVaaUnfiltered(event.ChainID.String(), event.Source)

	// Process the VAA
	p := ProcessSourceTxParams{
		TrackID:       event.TrackID,
		Timestamp:     event.Timestamp,
		ID:            event.ID,    // digest
		VaaId:         event.VaaID, // {chain/address/sequence}
		ChainId:       event.ChainID,
		Emitter:       event.EmitterAddress,
		Sequence:      event.Sequence,
		TxHash:        event.TxHash,
		Vaa:           event.Vaa,
		IsVaaSigned:   event.IsVaaSigned,
		Metrics:       c.metrics,
		Overwrite:     event.Overwrite, // avoid processing the same transaction twice
		Source:        event.Source,
		SentTimestamp: msg.SentTimestamp(),
	}
	_, err := ProcessSourceTx(ctx, c.logger, c.rpcpool, c.wormchainRpcPool, c.repository, &p, c.p2pNetwork, c.notionalCache)

	// add vaa processing duration metrics
	c.metrics.AddVaaProcessedDuration(uint16(event.ChainID), time.Since(start).Seconds())

	elapsedLog := zap.Uint64("elapsedTime", uint64(time.Since(start).Milliseconds()))
	// Log a message informing the processing status
	if errors.Is(err, chains.ErrChainNotSupported) {
		msg.Done()
		c.logger.Info("Skipping VAA - chain not supported",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			elapsedLog,
		)
	} else if errors.Is(err, ErrAlreadyProcessed) {
		msg.Done()
		c.logger.Warn("Origin message already processed - skipping",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			elapsedLog,
		)
	} else if err != nil {
		msg.Failed()
		c.logger.Error("Failed to process originTx",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			zap.Error(err),
			elapsedLog,
		)
	} else {
		msg.Done()
		c.logger.Info("Origin transaction processed successfully",
			zap.String("trackId", event.TrackID),
			zap.String("id", event.ID),
			elapsedLog,
		)
		c.metrics.IncOriginTxInserted(event.ChainID.String(), event.Source)
	}
}

func (c *Consumer) processTargetTx(ctx context.Context, msg queue.ConsumerMessage) {

	event := msg.Data()

	attr, ok := queue.GetAttributes[*queue.TargetChainAttributes](event)
	if !ok || attr == nil {
		msg.Failed()
		c.logger.Error("Failed to get attributes from message", zap.String("trackId", event.TrackID), zap.String("vaaId", event.ID))
		return
	}
	start := time.Now()

	// evm fee
	var evmFee *EvmFee
	if attr.GasUsed != nil && attr.EffectiveGasPrice != nil {
		evmFee = &EvmFee{
			GasUsed:           *attr.GasUsed,
			EffectiveGasPrice: *attr.EffectiveGasPrice,
		}
	}

	// solana fee
	var solanaFee *SolanaFee
	if attr.Fee != nil {
		solanaFee = &SolanaFee{
			Fee: *attr.Fee,
		}
	}

	// Process the VAA
	p := ProcessTargetTxParams{
		Source:         event.Source,
		TrackID:        event.TrackID,
		ID:             event.ID,    // digest
		VaaID:          event.VaaID, // {chain/address/sequence}
		ChainID:        event.ChainID,
		Emitter:        event.EmitterAddress,
		TxHash:         event.TxHash,
		BlockTimestamp: event.Timestamp,
		BlockHeight:    attr.BlockHeight,
		Method:         attr.Method,
		From:           attr.From,
		To:             attr.To,
		Status:         attr.Status,
		EvmFee:         evmFee,
		SolanaFee:      solanaFee,
		Metrics:        c.metrics,
		P2pNetwork:     c.p2pNetwork,
	}
	err := ProcessTargetTx(ctx, c.logger, c.repository, &p, c.notionalCache)

	elapsedLog := zap.Uint64("elapsedTime", uint64(time.Since(start).Milliseconds()))
	if err != nil {
		msg.Failed()
		c.logger.Error("Failed to process destinationTx",
			zap.String("trackId", event.TrackID),
			zap.String("vaaId", event.ID),
			zap.Error(err),
			elapsedLog,
		)
	} else {
		msg.Done()
		c.logger.Info("Destination transaction processed successfully",
			zap.String("trackId", event.TrackID),
			zap.String("id", event.ID),
			elapsedLog,
		)
	}
}
