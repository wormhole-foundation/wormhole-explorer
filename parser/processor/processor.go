package processor

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/parser/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"go.uber.org/zap"
)

const (
	portalTokenBridgeAppID         = "PORTAL_TOKEN_BRIDGE"
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
	metrics    *metrics.Metrics
	logger     *zap.Logger
}

func New(repository *parser.Repository, metrics *metrics.Metrics, logger *zap.Logger) *Processor {
	return &Processor{
		repository: repository,
		metrics:    metrics,
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

	p.logger.Info("Vaa save in repository", zap.String("id", vaaParsed.ID))

	if vaaParsed.AppID == portalTokenBridgeAppID {
		input, ok := vaaParsed.Result.(map[string]interface{})
		if ok {
			var result portalTokenBridgePayload
			err := mapstructure.Decode(input, &result)
			if err != nil {
				p.logger.Warn("Decoding map to payload struct", zap.String("id", vaaParsed.ID), zap.Error(err))
				return nil
			}
			if result.PayloadType == transferPayloadType || result.PayloadType == transferWithPayloadPayloadType {
				if result.Amount == nil || result.ToChainID == nil {
					p.logger.Warn("amount or toChain are empty", zap.String("id", vaaParsed.ID), zap.Any("payload", input))
					return nil
				}
				metric := &metrics.Volume{
					ChainSourceID:      vaaParsed.EmitterChain,
					ChainDestinationID: *result.ToChainID,
					Value:              *result.Amount,
					Timestamp:          vaaParsed.Timestamp,
					AppID:              vaaParsed.AppID,
				}
				err := p.metrics.PushVolume(ctx, metric)
				if err != nil {
					return err
				}
			}
		} else {
			p.logger.Warn("Casting parsed vaa to map", zap.String("id", vaaParsed.ID))
		}
	}

	return nil
}
