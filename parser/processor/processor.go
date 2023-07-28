package processor

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	parserAlert "github.com/wormhole-foundation/wormhole-explorer/parser/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Processor struct {
	parser     vaaPayloadParser.ParserVAAAPIClient
	repository *parser.Repository
	alert      alert.AlertClient
	metrics    metrics.Metrics
	logger     *zap.Logger
}

func New(parser vaaPayloadParser.ParserVAAAPIClient, repository *parser.Repository, alert alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger) *Processor {
	return &Processor{
		parser:     parser,
		repository: repository,
		alert:      alert,
		metrics:    metrics,
		logger:     logger,
	}
}

func (p *Processor) Process(ctx context.Context, vaaBytes []byte) (*parser.ParsedVaaUpdate, error) {
	// unmarshal vaa.
	vaa, err := sdk.Unmarshal(vaaBytes)
	if err != nil {
		return nil, err
	}

	// call vaa-payload-parser api to parse a VAA.
	chainID := uint16(vaa.EmitterChain)
	emitterAddress := vaa.EmitterAddress.String()
	sequence := fmt.Sprintf("%d", vaa.Sequence)

	p.metrics.IncVaaPayloadParserRequestCount(chainID)
	vaaParseResponse, err := p.parser.ParseVaaWithStandarizedProperties(vaa)
	if err != nil {
		// split metrics error not found and others errors.
		if errors.Is(err, vaaPayloadParser.ErrNotFound) {
			p.metrics.IncVaaPayloadParserNotFoundCount(chainID)
		} else {
			p.metrics.IncVaaPayloadParserErrorCount(chainID)
		}

		// if error is ErrInternalError or ErrCallEndpoint return error in order to retry.
		if errors.Is(err, vaaPayloadParser.ErrInternalError) || errors.Is(err, vaaPayloadParser.ErrCallEndpoint) {
			// send alert when exists and error calling vaa-payload-parser component.
			alertContext := alert.AlertContext{
				Details: map[string]string{
					"chainID":        vaa.EmitterChain.String(),
					"emitterAddress": emitterAddress,
					"sequence":       sequence,
				},
				Error: err,
			}
			p.alert.CreateAndSend(ctx, parserAlert.AlertKeyVaaPayloadParserError, alertContext)
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

	standardizedProperties := p.transformStandarizedProperties(vaa.MessageID(), vaaParseResponse.StandardizedProperties)

	// create ParsedVaaUpdate to upsert.
	now := time.Now()
	vaaParsed := parser.ParsedVaaUpdate{
		ID:                        vaa.MessageID(),
		EmitterChain:              vaa.EmitterChain,
		EmitterAddr:               emitterAddress,
		Sequence:                  sequence,
		AppIDs:                    standardizedProperties.AppIds,
		ParsedPayload:             vaaParseResponse.ParsedPayload,
		RawStandardizedProperties: vaaParseResponse.StandardizedProperties,
		StandardizedProperties:    standardizedProperties,
		Timestamp:                 vaa.Timestamp,
		UpdatedAt:                 &now,
	}

	err = p.repository.UpsertParsedVaa(ctx, vaaParsed)
	if err != nil {
		p.logger.Error("Error inserting vaa in repository",
			zap.String("id", vaaParsed.ID),
			zap.Error(err))
		// send alert when exists and error inserting parsed vaa.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"chainID":        vaa.EmitterChain.String(),
				"emitterAddress": emitterAddress,
				"sequence":       sequence,
				"appIDs":         strings.Join(standardizedProperties.AppIds, ", "),
			},
			Error: err}
		p.alert.CreateAndSend(ctx, parserAlert.AlertKeyInsertParsedVaaError, alertContext)
		return nil, err
	}
	p.metrics.IncVaaParsedInserted(chainID)

	p.logger.Info("parsed VAA was successfully persisted", zap.String("id", vaaParsed.ID))
	return &vaaParsed, nil
}

// transformStandarizedProperties transform amount and fee amount.
func (p *Processor) transformStandarizedProperties(vaaID string, sp vaaPayloadParser.StandardizedProperties) vaaPayloadParser.StandardizedProperties {
	// transform amount.
	amount := p.transformAmount(sp.TokenChain, sp.TokenAddress, sp.Amount, vaaID)
	// transform fee amount.
	feeAmount := p.transformAmount(sp.FeeChain, sp.FeeAddress, sp.Fee, vaaID)
	// create StandardizedProperties.
	return vaaPayloadParser.StandardizedProperties{
		AppIds:       sp.AppIds,
		FromChain:    sp.FromChain,
		FromAddress:  sp.FromAddress,
		ToChain:      sp.ToChain,
		ToAddress:    sp.ToAddress,
		TokenChain:   sp.TokenChain,
		TokenAddress: sp.TokenAddress,
		Amount:       amount,
		FeeAddress:   sp.FeeAddress,
		FeeChain:     sp.FeeChain,
		Fee:          feeAmount,
	}
}

// transformAmount transform amount and fee amount.
func (p *Processor) transformAmount(chainID sdk.ChainID, nativeAddress, amount, vaaID string) string {

	if chainID == sdk.ChainIDUnset || nativeAddress == "" || amount == "" {
		return ""
	}

	nativeHex, err := domain.DecodeNativeAddressToHex(sdk.ChainID(chainID), nativeAddress)
	if err != nil {
		p.logger.Warn("Native address cannot be transformed to hex",
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}

	addr, err := sdk.StringToAddress(nativeHex)
	if err != nil {
		p.logger.Warn("Address cannot be parsed",
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}
	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := domain.GetTokenByAddress(sdk.ChainID(chainID), addr.String())
	if !ok {
		p.logger.Warn("Token metadata not found",
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}

	bigAmount := new(big.Int)
	bigAmount, ok = bigAmount.SetString(amount, 10)
	if !ok {
		p.logger.Error("Cannot parse amount",
			zap.String("vaaId", vaaID),
			zap.String("amount", amount),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}

	if tokenMeta.Decimals < 8 {
		// factor = 10 ^ (8 - tokenMeta.Decimals)
		var factor big.Int
		factor.Exp(big.NewInt(10), big.NewInt(int64(8-tokenMeta.Decimals)), nil)

		bigAmount = bigAmount.Mul(bigAmount, &factor)
	}

	return bigAmount.String()
}

// createStandarizedProperties create a new StandardizedProperties with amount and fee amount transformed.
func createStandarizedProperties(m vaaPayloadParser.StandardizedProperties, amount, feeAmount, fromAddress, toAddress, tokenAddress, feeAddress string) vaaPayloadParser.StandardizedProperties {
	return vaaPayloadParser.StandardizedProperties{
		AppIds:       m.AppIds,
		FromChain:    m.FromChain,
		FromAddress:  fromAddress,
		ToChain:      m.ToChain,
		ToAddress:    toAddress,
		TokenChain:   m.TokenChain,
		TokenAddress: tokenAddress,
		Amount:       amount,
		FeeAddress:   feeAddress,
		FeeChain:     m.FeeChain,
		Fee:          feeAmount,
	}
}
