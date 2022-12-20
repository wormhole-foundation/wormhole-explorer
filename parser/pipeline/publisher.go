package pipeline

import (
	"context"
	"errors"

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

// Publish sends a signed VAA with parse configuration.
func (p *Publisher) Publish(e *watcher.Event) {
	// deserializes the binary representation of a VAA
	vaa, err := vaa.Unmarshal(e.Vaas)
	if err != nil {
		p.logger.Error("error Unmarshal vaa", zap.Error(err))
		return
	}

	// check exists vaa parser function by emitter chainID and emitterAddress
	vaaParser, err := p.repository.GetVaaParserFunction(context.TODO(), vaa.EmitterChain, vaa.EmitterAddress.String())
	if err != nil {
		if errors.Is(err, parser.ErrNotFound) {
			p.logger.Info("vaaParserFunction not found", zap.Uint16("chainID", uint16(vaa.EmitterChain)),
				zap.String("address", vaa.EmitterAddress.String()))
			return
		}
		p.logger.Error("can not get vaaParserFunction", zap.Error(err), zap.Uint16("chainID", uint16(vaa.EmitterChain)),
			zap.String("address", vaa.EmitterAddress.String()))
		return
	}

	// TODO: In V2:
	// We are going to add a in-memory cache for parser functions to avoid do a request per vaa to DB.
	// Wa are going to refresh the cache when a parser function is craeted/deleted/updated.

	// create a VaaEvent.
	event := queue.VaaEvent{
		ChainID:          vaa.EmitterChain,
		EmitterAddress:   vaa.EmitterAddress,
		Sequence:         vaa.Sequence,
		Vaa:              vaa.Payload,
		ParserFunctionID: vaaParser.ID,
	}

	// push messages to queue.
	err = p.pushFunc(context.TODO(), &event)
	if err != nil {
		p.logger.Error("can not push event to queue", zap.Error(err), zap.String("event", event.ID()))
	}
}
