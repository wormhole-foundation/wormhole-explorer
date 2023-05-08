package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"go.uber.org/zap"
)

const (
	transferPayloadType            = 1
	attestMetaPayloadType          = 2
	transferWithPayloadPayloadType = 3
)

type portalTokenBridgePayload struct {
	PayloadType int     `mapstructure:"payloadType"`
	Amount      *uint64 `mapstructure:"amount"`
	ToChainID   *uint16 `mapstructure:"toChain"`
}

type Processor struct {
	repository *parser.Repository
	logger     *zap.Logger
}

func New(repository *parser.Repository, logger *zap.Logger) *Processor {
	return &Processor{
		repository: repository,
		logger:     logger,
	}
}

func (p *Processor) Process(ctx context.Context, vaaParsed *parser.ParsedVaaUpdate) error {

	err := p.repository.UpsertParsedVaa(ctx, *vaaParsed)
	if err != nil {
		p.logger.Error("Error inserting vaa in repository",
			zap.String("id", vaaParsed.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("parsed VAA was successfully persisted", zap.String("id", vaaParsed.ID))
	return nil
}
