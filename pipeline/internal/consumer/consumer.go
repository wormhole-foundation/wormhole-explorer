package consumer

import (
	"context"
	"encoding/json"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/queue"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"time"
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
	txHash, err := c.postreSqlRepository.GetTxHash(ctx, event.ID)
	if err != nil {
		c.logger.Error("Error getting txHash", zap.Error(err), zap.String("vaaId", event.VaaId), zap.String("trackId", event.TrackID))
		msg.Failed()
		return
	}

	opTransaction := buildOperationTransaction(event, txHash)
	if event.ChainID != sdk.ChainIDWormchain && event.ChainID != sdk.ChainIDAptos && event.ChainID != sdk.ChainIDSolana {
		err = c.postreSqlRepository.CreateOperationTransaction(ctx, opTransaction)
		if err != nil {
			c.logger.Error("Error creating operation transaction", zap.Error(err), zap.String("vaaId", event.VaaId), zap.String("trackId", event.TrackID))
			msg.Failed()
			return
		}
	}

	err = c.publishVaa(ctx, &opTransaction)
	if err != nil {
		c.logger.Error("Error publishing VAA", zap.Error(err), zap.String("vaaId", event.VaaId), zap.String("trackId", event.TrackID))
		c.metrics.IncVaaFailedProcessing(event.ChainID, "")
		msg.Failed()
		return
	}

	c.logger.Debug("VAA processed", zap.String("vaaId", event.VaaId), zap.String("trackId", event.TrackID))
	c.metrics.IncVaaSendNotificationFromGossipSQS(event.ChainID)
	msg.Done()
}

func buildOperationTransaction(event *queue.Event, txHash string) operationTransaction {
	return operationTransaction{
		ChainID:          event.ChainID,
		TxHash:           txHash,
		Type:             event.Type,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		AttestationVaaID: event.ID,
		VaaID:            event.VaaId,
		FromAddress:      &event.EmitterAddress,
		Timestamp:        event.Timestamp,
	}
}

type operationTransaction struct {
	ChainID          sdk.ChainID     `json:"chain_id"`
	TxHash           string          `json:"tx_hash"`
	Type             queue.EventType `json:"type"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	AttestationVaaID string          `json:"attestation_vaas_id"`
	VaaID            string          `json:"vaaId"`
	FromAddress      *string         `json:"from_address"`
	Timestamp        *time.Time      `json:"timestamp"`
}

func (o *operationTransaction) GetGroupID() string {
	return o.AttestationVaaID
}

func (o *operationTransaction) GetDeduplicationID() string {
	return o.AttestationVaaID
}

func (o *operationTransaction) GetChainID() sdk.ChainID {
	return o.ChainID
}

func (o *operationTransaction) Body() ([]byte, error) {
	return json.Marshal(o)
}
