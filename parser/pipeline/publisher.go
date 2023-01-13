package pipeline

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"github.com/wormhole-foundation/wormhole-explorer/parser/watcher"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Publisher definition.
type Publisher struct {
	logger     *zap.Logger
	repository *parser.Repository
	pushFunc   queue.VAAPushFunc
}

// NewPublisher creates a new publisher for vaa with parse configuration.
func NewPublisher(logger *zap.Logger, repository *parser.Repository, pushFunc queue.VAAPushFunc) *Publisher {
	return &Publisher{logger: logger, repository: repository, pushFunc: pushFunc}
}

// Publish sends a VaaEvent for the vaa that has parse configuration defined.
func (p *Publisher) Publish(e *watcher.Event) {
	// deserializes the binary representation of a VAA
	vaa, err := vaa.Unmarshal(e.Vaas)
	if err != nil {
		p.logger.Error("error Unmarshal vaa", zap.Error(err))
		return
	}

	// V2 Get chainID/address that have parser function defined and add to sqs only that vaa.

	// create a VaaEvent.
	event := queue.VaaEvent{
		ChainID:        uint16(vaa.EmitterChain),
		EmitterAddress: vaa.EmitterAddress.String(),
		Sequence:       vaa.Sequence,
		Vaa:            vaa.Payload,
	}

	// push messages to queue.
	err = p.pushFunc(context.TODO(), &event)
	if err != nil {
		p.logger.Error("can not push event to queue", zap.Error(err), zap.String("event", event.ID()))
	}
}
