package pipeline

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/watcher"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Publisher definition.
type Publisher struct {
	logger        *zap.Logger
	pushFunc      topic.PushFunc
	repository    *Repository
	p2pNetwork    string
	txHashHandler *TxHashHandler
	metrics       metrics.Metrics
}

// NewPublisher creates a new publisher for vaa with parse configuration.
func NewPublisher(pushFunc topic.PushFunc, metrics metrics.Metrics, repository *Repository, p2pNetwork string, txHashHandler *TxHashHandler, logger *zap.Logger) *Publisher {
	return &Publisher{
		logger:        logger,
		repository:    repository,
		pushFunc:      pushFunc,
		p2pNetwork:    p2pNetwork,
		txHashHandler: txHashHandler,
		metrics:       metrics,
	}
}

// Publish sends a Event for the vaa that has parse configuration defined.
func (p *Publisher) Publish(ctx context.Context, e *watcher.Event) {

	// create a Event.
	event := topic.Event{
		ID:               e.ID,
		ChainID:          e.ChainID,
		EmitterAddress:   e.EmitterAddress,
		Sequence:         e.Sequence,
		GuardianSetIndex: e.GuardianSetIndex,
		Vaa:              e.Vaa,
		IndexedAt:        e.IndexedAt,
		Timestamp:        e.Timestamp,
		UpdatedAt:        e.UpdatedAt,
		TxHash:           e.TxHash,
		Version:          e.Version,
		Revision:         e.Revision,
		Hash:             e.Hash,
	}

	// In some scenarios the fly component that inserts the VAA documents does not have the txhash field available,
	// since this field does not arrive in the gossip network messages of type vaa, but arrives in the messages
	// of type observation and there may be a race condition between the processing of observations and the vaa.
	// For this reason, an attempt is made at this point to complete this vaa with the txhash.
	if event.TxHash == "" {
		// discard pyth messages
		isPyth := vaa.ChainIDPythNet == vaa.ChainID(e.ChainID)
		if !isPyth {
			// increment the metric for the number of vaas without txhash
			p.metrics.IncVaaWithoutTxHash(e.ChainID)
			// add the event to the txhash handler.
			// the handler will try to get the txhash for the vaa
			// and publish the event with the txhash.
			p.txHashHandler.AddVaaFixItem(event)
			return
		}
	}

	// push messages to topic.
	err := p.pushFunc(ctx, &event)
	if err != nil {
		p.logger.Error("can not push event to topic", zap.Error(err), zap.String("event", event.ID))
	}
}
