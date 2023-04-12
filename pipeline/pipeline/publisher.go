package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/watcher"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Publisher definition.
type Publisher struct {
	logger     *zap.Logger
	pushFunc   topic.PushFunc
	repository *Repository
	p2pNetwork string
}

// NewPublisher creates a new publisher for vaa with parse configuration.
func NewPublisher(pushFunc topic.PushFunc, repository *Repository, p2pNetwork string, logger *zap.Logger) *Publisher {
	return &Publisher{logger: logger, repository: repository, pushFunc: pushFunc, p2pNetwork: p2pNetwork}
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
	}

	// In some scenarios the fly component that inserts the VAA documents does not have the txhash field available,
	// since this field does not arrive in the gossip network messages of type vaa, but arrives in the messages
	// of type observation and there may be a race condition between the processing of observations and the vaa.
	// For this reason, an attempt is made at this point to complete this vaa with the txhash.
	if event.TxHash == "" {
		// discard pyth messages
		isPyth := domain.P2pMainNet == p.p2pNetwork && vaa.ChainIDPythNet == vaa.ChainID(e.ChainID)
		if !isPyth {
			// retry 3 times with 2 seconds delay fixing the vaa with empty txhash.
			txHash, err := Retry(p.handleEmptyVaaTxHash, 3, 2*time.Second)(ctx, e.ID)
			if err != nil {
				p.logger.Error("can not get txhash for vaa", zap.Error(err), zap.String("event", event.ID))
			}
			event.TxHash = txHash
		}
	}

	// push messages to topic.
	err := p.pushFunc(ctx, &event)
	if err != nil {
		p.logger.Error("can not push event to topic", zap.Error(err), zap.String("event", event.ID))
	}
}

// handleEmptyVaaTxHash tries to get the txhash for the vaa with the given id.
func (p *Publisher) handleEmptyVaaTxHash(ctx context.Context, id string) (string, error) {
	vaaIdTxHash, err := p.repository.GetVaaIdTxHash(ctx, id)
	if err != nil {
		return "", err
	}

	if vaaIdTxHash.TxHash == "" {
		return "", fmt.Errorf("txhash for vaa (%s) is empty", id)
	}

	err = p.repository.UpdateVaaDocTxHash(ctx, id, vaaIdTxHash.TxHash)
	if err != nil {
		return "", err
	}
	return vaaIdTxHash.TxHash, nil
}

// RetryFn is a function that can be retried.
type RetryFn func(ctx context.Context, id string) (string, error)

// Retry retries a function.
func Retry(retryFn RetryFn, retries int, delay time.Duration) RetryFn {
	return func(ctx context.Context, id string) (string, error) {
		var err error
		var txHash string
		for i := 0; i <= retries; i++ {
			txHash, err = retryFn(ctx, id)
			if err == nil {
				return txHash, nil
			}
			time.Sleep(delay)
		}
		return txHash, err
	}
}
