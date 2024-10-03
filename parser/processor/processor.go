package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/tidwall/gjson"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	parserAlert "github.com/wormhole-foundation/wormhole-explorer/parser/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Processor struct {
	parser             vaaPayloadParser.ParserVAAAPIClient
	dbMode             string
	mongoRepository    *parser.Repository
	postgresRepository *parser.PostgresRepository
	alert              alert.AlertClient
	metrics            metrics.Metrics
	tokenProvider      *domain.TokenProvider
	logger             *zap.Logger
}

func New(parser vaaPayloadParser.ParserVAAAPIClient, dbMode string, mongoRepository *parser.Repository,
	postgresRepository *parser.PostgresRepository, alert alert.AlertClient, metrics metrics.Metrics,
	tokenProvider *domain.TokenProvider, logger *zap.Logger) *Processor {
	return &Processor{
		parser:             parser,
		dbMode:             dbMode,
		mongoRepository:    mongoRepository,
		postgresRepository: postgresRepository,
		alert:              alert,
		metrics:            metrics,
		tokenProvider:      tokenProvider,
		logger:             logger,
	}
}

func (p *Processor) Process(ctx context.Context, params *Params) (*parser.ParsedVaaUpdate, error) {
	// create logger with trackId.
	logger := p.logger.With(zap.String("trackId", params.TrackID))

	// unmarshal vaa.
	vaa, err := sdk.Unmarshal(params.Vaa)
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
					"trackID":        params.TrackID,
					"chainID":        vaa.EmitterChain.String(),
					"emitterAddress": emitterAddress,
					"sequence":       sequence,
				},
				Error: err,
			}
			p.alert.CreateAndSend(ctx, parserAlert.AlertKeyVaaPayloadParserError, alertContext)
			return nil, err
		}

		logger.Info("VAA cannot be parsed", zap.Error(err),
			zap.String("trackId", params.TrackID),
			zap.Uint16("chainId", chainID),
			zap.String("address", emitterAddress),
			zap.String("sequence", sequence))
		return nil, nil
	}
	p.metrics.IncVaaPayloadParserSuccessCount(chainID)
	p.metrics.IncVaaParsed(chainID)

	standardizedProperties := p.transformStandarizedProperties(
		params.TrackID, vaa.MessageID(), vaaParseResponse.StandardizedProperties)
	now := time.Now()
	err = p.upserteVaaParse(ctx, vaa, vaaParseResponse,
		standardizedProperties, now, logger)
	if err != nil {
		logger.Error("Error inserting parsed vaa",
			zap.String("trackId", params.TrackID),
			zap.String("id", utils.NormalizeHex(vaa.HexDigest())),
			zap.String("vaaId", vaa.MessageID()),
			zap.Error(err))

		alertContext := alert.AlertContext{
			Details: map[string]string{
				"trackID":        params.TrackID,
				"chainID":        vaa.EmitterChain.String(),
				"emitterAddress": emitterAddress,
				"sequence":       sequence,
			},
			Error: err}
		p.alert.CreateAndSend(ctx, parserAlert.AlertKeyInsertParsedVaaError, alertContext)
		return nil, err
	}

	p.metrics.IncVaaParsedInserted(chainID)
	logger.Info("parsed VAA was successfully persisted",
		zap.String("trackId", params.TrackID),
		zap.String("id", utils.NormalizeHex(vaa.HexDigest())),
		zap.String("vaaId", vaa.MessageID()))

	return &parser.ParsedVaaUpdate{
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
	}, nil
}

func (p *Processor) upserteVaaParse(ctx context.Context, vaa *sdk.VAA,
	vaaParseResponse *vaaPayloadParser.ParseVaaWithStandarizedPropertiesdResponse,
	standardizedProperties vaaPayloadParser.StandardizedProperties, now time.Time,
	logger *zap.Logger) error {

	emitterAddress := vaa.EmitterAddress.String()
	sequence := fmt.Sprintf("%d", vaa.Sequence)

	switch p.dbMode {
	case config.DbLayerMongo:
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

		err := p.mongoRepository.UpsertParsedVaa(ctx, vaaParsed)
		if err == nil {
			p.metrics.IncParseVaaInserted(uint16(vaa.EmitterChain))
		}
		return err
	case config.DbLayerPostgres:
		attestationVaaProperties, err := buildAttestationVaaProperties(
			vaa, standardizedProperties, vaaParseResponse, logger)
		if err != nil {
			return err
		}

		err = p.postgresRepository.UpsertAttestationVaaProperties(ctx, attestationVaaProperties)
		if err == nil {
			p.metrics.IncVaaAttestationPropertiesInserted(uint16(vaa.EmitterChain))
		}
		return err
	case config.DbLayerDual:
		// upsert mongo
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

		err := p.mongoRepository.UpsertParsedVaa(ctx, vaaParsed)
		if err != nil {
			p.logger.Error("Error inserting vaa in mongo repository",
				zap.String("id", utils.NormalizeHex(vaa.HexDigest())),
				zap.Error(err))
			return err
		}
		p.metrics.IncParseVaaInserted(uint16(vaa.EmitterChain))

		// upsert postgres
		attestationVaaProperties, err := buildAttestationVaaProperties(
			vaa, standardizedProperties, vaaParseResponse, p.logger)
		if err != nil {
			return err
		}

		err = p.postgresRepository.UpsertAttestationVaaProperties(ctx, attestationVaaProperties)
		if err == nil {
			p.metrics.IncVaaAttestationPropertiesInserted(uint16(vaa.EmitterChain))
		}
		return err
	default:
		return errors.New("invalid db mode")
	}
}

// buildAttestationVaaProperties build AttestationVaaProperties from
// vaa, standardizedProperties and vaaParseResponse.
func buildAttestationVaaProperties(
	vaa *sdk.VAA,
	standardizedProperties vaaPayloadParser.StandardizedProperties,
	vaaParseResponse *vaaPayloadParser.ParseVaaWithStandarizedPropertiesdResponse,
	logger *zap.Logger) (parser.AttestationVaaProperties, error) {

	// get raw payload.
	var rawPayload *json.RawMessage
	var payloadType *int
	if parsedPayload := vaaParseResponse.ParsedPayload; parsedPayload != nil {
		jsonPayload, err := json.Marshal(parsedPayload)
		if err != nil {
			logger.Error("Error marshalling parsed payload",
				zap.String("vaaId", vaa.MessageID()),
				zap.Error(err))
			return parser.AttestationVaaProperties{}, err
		}
		rawPayload = (*json.RawMessage)(&jsonPayload)
		payloadTypeValue := gjson.Get(string(jsonPayload), "payloadType")
		if payloadTypeValue.Exists() {
			value := int(payloadTypeValue.Int())
			if value > 0 {
				payloadType = &value
			}
		}
	}

	// get raw standardized fields.
	var rawStandardFields *json.RawMessage
	jsonProperties, err := json.Marshal(standardizedProperties)
	if err != nil {
		logger.Error("Error marshalling standardized properties",
			zap.String("vaaId", vaa.MessageID()),
			zap.Error(err))
		return parser.AttestationVaaProperties{}, err
	}
	rawStandardFields = (*json.RawMessage)(&jsonProperties)

	// normalize fromChainID.
	var fromChainID *sdk.ChainID
	StandardizedPropsFromChainID := sdk.ChainID(standardizedProperties.FromChain)
	if StandardizedPropsFromChainID != sdk.ChainIDUnset {
		fromChainID = &StandardizedPropsFromChainID
	}

	// normalize fromAddress.
	var fromAddress *string
	if standardizedProperties.FromAddress != "" {
		normalizedAddress := domain.NormalizeAddressByChainId(
			StandardizedPropsFromChainID, standardizedProperties.FromAddress)
		fromAddress = &normalizedAddress
	}

	// normalize toChainID.
	var toChainID *sdk.ChainID
	StandardizedPropsToChainID := sdk.ChainID(standardizedProperties.ToChain)
	if StandardizedPropsToChainID != sdk.ChainIDUnset {
		toChainID = &StandardizedPropsToChainID
	}

	// normalize toAddress.
	var toAddress *string
	if standardizedProperties.ToAddress != "" {
		normalizedAddress := domain.NormalizeAddressByChainId(
			StandardizedPropsToChainID, standardizedProperties.ToAddress)
		toAddress = &normalizedAddress
	}

	// normalize tokenChainID.
	var tokenChainID *sdk.ChainID
	StandardizedPropsTokenChainID := sdk.ChainID(standardizedProperties.TokenChain)
	if StandardizedPropsTokenChainID != sdk.ChainIDUnset {
		tokenChainID = &StandardizedPropsTokenChainID
	}

	// normalize tokenAddress.
	var tokenAddress *string
	if standardizedProperties.TokenAddress != "" {
		normalizedAddress := domain.NormalizeAddressByChainId(
			StandardizedPropsTokenChainID, standardizedProperties.TokenAddress)
		tokenAddress = &normalizedAddress
	}

	// normalize feeChainID.
	var feeChainID *sdk.ChainID
	StandardizedPropsFeeChainID := sdk.ChainID(standardizedProperties.FeeChain)
	if StandardizedPropsFeeChainID != sdk.ChainIDUnset {
		feeChainID = &StandardizedPropsFeeChainID
	}

	// normalize feeAddress.
	var feeAddress *string
	if standardizedProperties.FeeAddress != "" {
		normalizedAddress := domain.NormalizeAddressByChainId(
			StandardizedPropsFeeChainID, standardizedProperties.FeeAddress)
		feeAddress = &normalizedAddress
	}

	// get bit.Int amount
	var amount *big.Int
	if standardizedProperties.Amount != "" {
		amount = new(big.Int)
		if _, ok := amount.SetString(standardizedProperties.Amount, 10); !ok {
			logger.Error("Cannot parse amount",
				zap.String("vaaId", vaa.MessageID()),
				zap.String("amount", standardizedProperties.Amount))
			return parser.AttestationVaaProperties{}, errors.New("cannot parse amount")
		}
	}

	// get bit.Int fee
	var fee *big.Int
	if standardizedProperties.Fee != "" {
		fee = new(big.Int)
		if _, ok := fee.SetString(standardizedProperties.Fee, 10); !ok {
			logger.Error("Cannot parse fee",
				zap.String("vaaId", vaa.MessageID()),
				zap.String("fee", standardizedProperties.Fee))
			return parser.AttestationVaaProperties{}, errors.New("cannot parse fee")
		}
	}

	return parser.AttestationVaaProperties{
		ID:                utils.NormalizeHex(vaa.HexDigest()),
		VaaID:             vaa.MessageID(),
		AppID:             standardizedProperties.AppIds,
		Payload:           rawPayload,
		PayloadType:       payloadType,
		RawStandardFields: rawStandardFields,
		FromChainID:       fromChainID,
		FromAddress:       fromAddress,
		ToChainID:         toChainID,
		ToAddress:         toAddress,
		TokenChainID:      tokenChainID,
		TokenAddress:      tokenAddress,
		Amount:            amount,
		FeeChainID:        feeChainID,
		FeeAddress:        feeAddress,
		Fee:               fee,
		Timestamp:         vaa.Timestamp,
	}, nil
}

// transformStandarizedProperties transform amount and fee amount.
func (p *Processor) transformStandarizedProperties(trackID, vaaID string, sp vaaPayloadParser.StandardizedProperties) vaaPayloadParser.StandardizedProperties {
	// transform amount.
	amount := p.transformAmount(sp.TokenChain, trackID, sp.TokenAddress, sp.Amount, vaaID)
	// transform fee amount.
	feeAmount := p.transformAmount(sp.FeeChain, trackID, sp.FeeAddress, sp.Fee, vaaID)
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
func (p *Processor) transformAmount(chainID sdk.ChainID, trackID, nativeAddress, amount, vaaID string) string {

	if chainID == sdk.ChainIDUnset || nativeAddress == "" || amount == "" {
		return ""
	}

	nativeHex, err := domain.DecodeNativeAddressToHex(sdk.ChainID(chainID), nativeAddress)
	if err != nil {
		p.logger.Warn("Native address cannot be transformed to hex",
			zap.String("trackId", trackID),
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}

	addr, err := sdk.StringToAddress(nativeHex)
	if err != nil {
		p.logger.Warn("Address cannot be parsed",
			zap.String("trackId", trackID),
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}
	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := p.tokenProvider.GetTokenByAddress(sdk.ChainID(chainID), addr.String())
	if !ok {
		p.logger.Warn("Token metadata not found",
			zap.String("trackId", trackID),
			zap.String("vaaId", vaaID),
			zap.String("nativeAddress", nativeAddress),
			zap.Uint16("chain", uint16(chainID)))
		return ""
	}

	bigAmount := new(big.Int)
	bigAmount, ok = bigAmount.SetString(amount, 10)
	if !ok {
		p.logger.Error("Cannot parse amount",
			zap.String("trackId", trackID),
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
