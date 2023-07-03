package processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Processor struct {
	parser     parser.ParserVAAAPIClient
	repository *parser.Repository
	metrics    metrics.Metrics
	logger     *zap.Logger
}

func New(parser parser.ParserVAAAPIClient, repository *parser.Repository, metrics metrics.Metrics, logger *zap.Logger) *Processor {
	return &Processor{
		parser:     parser,
		repository: repository,
		metrics:    metrics,
		logger:     logger,
	}
}

func (p *Processor) Process(ctx context.Context, vaaBytes []byte) (*parser.ParsedVaaUpdate, error) {
	// unmarshal vaa.
	vaa, err := vaa.Unmarshal(vaaBytes)
	if err != nil {
		return nil, err
	}

	// call vaa-payload-parser api to parse a VAA.
	chainID := uint16(vaa.EmitterChain)
	emitterAddress := vaa.EmitterAddress.String()
	sequence := fmt.Sprintf("%d", vaa.Sequence)

	p.metrics.IncVaaPayloadParserRequestCount(chainID)
	vaaParseResponse, err := p.parser.Parse(chainID, emitterAddress, sequence, vaa.Payload)
	if err != nil {
		// split metrics error not found and others errors.
		if errors.Is(err, parser.ErrNotFound) {
			p.metrics.IncVaaPayloadParserNotFoundCount(chainID)
		} else {
			p.metrics.IncVaaPayloadParserErrorCount(chainID)
		}

		// if error is ErrInternalError or ErrCallEndpoint return error in order to retry.
		if errors.Is(err, parser.ErrInternalError) || errors.Is(err, parser.ErrCallEndpoint) {
			return nil, err
		}

		p.logger.Info("VAA cannot be parsed", zap.Error(err),
			zap.Uint16("chainID", chainID),
			zap.String("address", emitterAddress),
			zap.String("sequence", sequence))
		return nil, nil
	}
	p.metrics.IncVaaPayloadParserSuccessCount(chainID)
	p.metrics.IncVaaParsed(chainID)

	// create ParsedVaaUpdate to upsert.
	now := time.Now()
	vaaParsed := parser.ParsedVaaUpdate{
		ID:           vaa.MessageID(),
		EmitterChain: chainID,
		EmitterAddr:  emitterAddress,
		Sequence:     sequence,
		AppID:        vaaParseResponse.AppID,
		Result:       vaaParseResponse.Result,
		Timestamp:    vaa.Timestamp,
		UpdatedAt:    &now,
	}

	err = p.repository.UpsertParsedVaa(ctx, vaaParsed)
	if err != nil {
		p.logger.Error("Error inserting vaa in repository",
			zap.String("id", vaaParsed.ID),
			zap.Error(err))
		return nil, err
	}
	p.metrics.IncVaaParsedInserted(chainID)

	p.logger.Info("parsed VAA was successfully persisted", zap.String("id", vaaParsed.ID))
	return &vaaParsed, nil
}
