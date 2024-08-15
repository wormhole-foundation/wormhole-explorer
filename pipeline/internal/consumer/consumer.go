package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/queue"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	postreSqlRepository PostreSqlRepository
	logger              *zap.Logger
	publishVaa          topic.PushFunc
	metrics             metrics.Metrics
	workersSize         int
}

// New creates a new vaa consumer.
func New(
	postreSqlRepository PostreSqlRepository,
	logger *zap.Logger,
	publishVaa topic.PushFunc,
	metrics metrics.Metrics,
	workersSize int,
) *Consumer {

	c := Consumer{
		postreSqlRepository: postreSqlRepository,
		logger:              logger,
		publishVaa:          publishVaa,
		metrics:             metrics,
		workersSize:         workersSize,
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context, consumeFunc queue.ConsumeFunc) {
	ch := consumeFunc(ctx)
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
			c.processVaaEvent(ctx, msg)
		}
	}
}

func (c *Consumer) processVaaEvent(ctx context.Context, msg queue.ConsumerMessage) {

	event := msg.Data()

	var txHash string
	var err error

	if event.EmitterChainID != sdk.ChainIDPythNet {
		txHash, err = c.postreSqlRepository.GetTxHash(ctx, event.ID)
		if err != nil {
			c.logger.Error("Error getting txHash", zap.Error(err),
				zap.String("attestation_vaa_id", event.ID),
				zap.String("trackId", event.TrackID))
			msg.Failed()
			return
		}

		opTransaction := buildOperationTransaction(event, txHash)
		if event.EmitterChainID != sdk.ChainIDWormchain && event.EmitterChainID != sdk.ChainIDAptos && event.EmitterChainID != sdk.ChainIDSolana {
			err = c.postreSqlRepository.CreateOperationTransaction(ctx, opTransaction)
			if err != nil {
				c.logger.Error("Error creating operation transaction", zap.Error(err),
					zap.String("attestation_vaa_id", event.ID),
					zap.String("trackId", event.TrackID))
			}
		}
	}

	pipelineEvent := topic.Event{
		ID:               event.VaaID,
		ChainID:          event.EmitterChainID,
		EmitterAddress:   event.EmitterAddress,
		Sequence:         fmt.Sprintf("%d", event.Sequence),
		GuardianSetIndex: event.GuardianSetIndex,
		Vaa:              event.Vaa,
		Timestamp:        &event.Timestamp,
		TxHash:           txHash,
		Version:          uint16(event.Version),
		Digest:           event.ID,
		Overwrite:        false, // each vaa has a unique digest in the pipeline processing.
	}

	err = c.publishVaa(ctx, pipelineEvent)
	if err != nil {
		c.logger.Error("Error publishing VAA", zap.Error(err),
			zap.String("id", event.ID),
			zap.String("trackId", event.TrackID))
		c.metrics.IncVaaFailedProcessing(event.EmitterChainID, "")
		msg.Failed()
		return
	}

	c.logger.Debug("VAA processed",
		zap.String("id", event.ID),
		zap.String("trackId", event.TrackID))
	c.metrics.IncVaaSendNotificationFromGossipSQS(event.EmitterChainID)
	msg.Done()
}

func buildOperationTransaction(event *queue.Event, txHash string) OperationTransaction {
	now := time.Now()
	return OperationTransaction{
		ChainID:          event.EmitterChainID,
		TxHash:           txHash,
		Type:             "source-tx",
		CreatedAt:        now,
		UpdatedAt:        now,
		AttestationVaaID: event.ID,
		VaaID:            event.VaaID,
		FromAddress:      &event.EmitterAddress,
		Timestamp:        &event.Timestamp,
	}
}

type OperationTransaction struct {
	ChainID          sdk.ChainID `json:"chain_id"`
	TxHash           string      `json:"tx_hash"`
	Type             string      `json:"type"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	AttestationVaaID string      `json:"attestation_vaas_id"`
	VaaID            string      `json:"vaaId"`
	FromAddress      *string     `json:"from_address"`
	Timestamp        *time.Time  `json:"timestamp"`
}
