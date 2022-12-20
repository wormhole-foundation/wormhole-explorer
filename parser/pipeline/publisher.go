package pipeline

import (
	"context"
	"errors"
	"fmt"

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
}

// NewPublisher creates a new publisher for vaa with parse configuration.
func NewPublisher(logger *zap.Logger, repository *parser.Repository) *Publisher {
	return &Publisher{logger: logger, repository: repository}
}

// Publish sends a signed VAA with parse configuration.
func (p *Publisher) Publish(e *watcher.Event) {
	// deserializes the binary representation of a VAA
	vaa, err := vaa.Unmarshal(e.Vaas)
	if err != nil {
		p.logger.Error("error Unmarshal vaa", zap.Error(err))
		return
	}

	// TODO delete this:
	// var chainID vaa.ChainID = 2
	// emitterAddress := "000000000000000000000000c63e43e2f09537a2b07fba1e02c6f4163a956525"

	// check exists vaa parser function by emitter chainID and emitterAddress
	vaaParser, err := p.repository.GetVaaParserFunction(context.TODO(), vaa.EmitterChain, vaa.EmitterAddress.String())
	if err != nil {
		if errors.Is(err, parser.ErrNotFound) {
			p.logger.Info("vaaParserFunction not found", zap.Uint16("chainID", uint16(vaa.EmitterChain)),
				zap.String("address", vaa.EmitterAddress.String()))
			return
		}
		p.logger.Error("error getting vaaParserFunction", zap.Error(err), zap.Uint16("chainID", uint16(vaa.EmitterChain)),
			zap.String("address", vaa.EmitterAddress.String()))
		return
	}

	// TODO: In V2:
	// We are going to add a in-memory cache for parser functions to avoid do a request per vaa to DB.
	// Wa are going to refresh the cache when a parser function is craeted/deleted/updated.

	// create a VaaEvent
	vaaEvent := queue.VaaEvent{
		ChainID:          vaa.EmitterChain,
		EmitterAddress:   vaa.EmitterAddress,
		Sequence:         vaa.Sequence,
		Vaa:              vaa.Payload,
		ParserFunctionID: vaaParser.ID,
	}

	// push message to sqs.
	fmt.Println(vaaParser)
}
