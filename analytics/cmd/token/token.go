package token

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type UnknownTokenErr struct {
	detail string
}

func (e *UnknownTokenErr) Error() string {
	return fmt.Sprintf("unknown token. %s", e.detail)
}

func IsUnknownTokenErr(err error) bool {
	switch err.(type) {
	case *UnknownTokenErr:
		return true
	default:
		return false
	}
}

type TransferredToken struct {
	AppId        string
	FromChain    sdk.ChainID
	ToChain      sdk.ChainID
	TokenAddress sdk.Address
	TokenChain   sdk.ChainID
	Amount       *big.Int
}

func (t *TransferredToken) Clone() *TransferredToken {
	if t == nil {
		return nil
	}
	var amount *big.Int
	if t.Amount != nil {
		amount = new(big.Int).Set(t.Amount)
	}
	return &TransferredToken{
		AppId:        t.AppId,
		FromChain:    t.FromChain,
		ToChain:      t.ToChain,
		TokenAddress: t.TokenAddress,
		TokenChain:   t.TokenChain,
		Amount:       amount,
	}
}

type GetTransferredTokenByVaa func(context.Context, *sdk.VAA) (*TransferredToken, error)

type TokenResolver struct {
	client parser.ParserVAAAPIClient
	logger *zap.Logger
}

func NewTokenResolver(client parser.ParserVAAAPIClient, logger *zap.Logger) *TokenResolver {
	return &TokenResolver{
		client: client,
		logger: logger,
	}
}

func (r *TokenResolver) GetTransferredTokenByVaa(ctx context.Context, vaa *sdk.VAA) (*TransferredToken, error) {

	// Ignore PythNet VAAs
	if vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil, nil
	}

	// Parse the VAA with standarized properties
	result, err := r.client.ParseVaaWithStandarizedProperties(vaa)
	if err != nil {
		if errors.Is(err, parser.ErrUnprocessableEntity) {
			r.logger.Debug("Parsing vaa with standarized properties resulted in unprocessable entity",
				zap.String("vaaId", vaa.MessageID()))
			return nil, nil
		}
		r.logger.Error("Parsing vaa with standarized properties",
			zap.String("vaaId", vaa.MessageID()),
			zap.Error(err))
		return nil, err
	}

	if result == nil {
		r.logger.Error("VAA with standarized properties is empty",
			zap.String("vaaId", vaa.MessageID()),
			zap.Error(err))
		return nil, nil
	}

	token, err := createToken(result, vaa.EmitterChain)
	if err != nil {
		r.logger.Debug("Creating transferred token",
			zap.String("vaaId", vaa.MessageID()),
			zap.Error(err))
		return nil, &UnknownTokenErr{detail: err.Error()}
	}

	return token, err
}

func createToken(p *parser.ParseVaaWithStandarizedPropertiesdResponse, emitterChain sdk.ChainID) (*TransferredToken, error) {

	if !domain.ChainIdIsValid(p.StandardizedProperties.TokenChain) {
		return nil, fmt.Errorf("tokenChain is invalid: %d", p.StandardizedProperties.TokenChain)
	}

	if !domain.ChainIdIsValid(p.StandardizedProperties.ToChain) {
		return nil, fmt.Errorf("toChain is invalid: %d", p.StandardizedProperties.ToChain)
	}

	if !domain.ChainIdIsValid(emitterChain) {
		return nil, fmt.Errorf("emitterChain is invalid: %d", emitterChain)
	}

	if p.StandardizedProperties.TokenAddress == "" {
		return nil, errors.New("tokenAddress is empty")
	}

	if p.StandardizedProperties.Amount == "" {
		return nil, errors.New("amount is empty")
	}

	addressHex, err := domain.DecodeNativeAddressToHex(p.StandardizedProperties.TokenChain, p.StandardizedProperties.TokenAddress)
	if err != nil {
		tokenParsedPayload, err := parseTokenPayload(p.ParsedPayload)
		if err != nil {
			return nil, fmt.Errorf("cannot decode token with tokenChain [%d] tokenAddress [%s] to hex. %v",
				p.StandardizedProperties.TokenChain, p.StandardizedProperties.TokenAddress, err)
		} else {
			addressHex = *tokenParsedPayload.TokenAddress
		}
	}

	address, err := sdk.StringToAddress(addressHex)
	if err != nil {
		return nil, err
	}

	n := new(big.Int)
	n, ok := n.SetString(p.StandardizedProperties.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("amount [%s] is not a number", p.StandardizedProperties.Amount)
	}

	appId := domain.AppIdUnkonwn
	if len(p.StandardizedProperties.AppIds) > 0 {
		appId = p.StandardizedProperties.AppIds[0]
	}

	return &TransferredToken{
		AppId:        appId,
		FromChain:    emitterChain,
		ToChain:      p.StandardizedProperties.ToChain,
		TokenAddress: address,
		TokenChain:   p.StandardizedProperties.TokenChain,
		Amount:       n,
	}, nil
}

func parseTokenPayload(parsedPayload any) (*tokenParsedPayload, error) {
	if parsedPayload == nil {
		return nil, fmt.Errorf("parsedPayload is nil")
	}
	var result *tokenParsedPayload
	err := mapstructure.Decode(parsedPayload, &result)
	if err != nil {
		return nil, fmt.Errorf("parsedPayload can not decode %v", err)
	}

	if result.TokenAddress == nil {
		return nil, fmt.Errorf("tokenAddress in parsedPayload is nil")

	}
	if result.TokenChain == nil {
		return nil, fmt.Errorf("tokenChain in parsedPayload is nil")
	}
	return result, nil
}

type tokenParsedPayload struct {
	TokenAddress *string `json:"tokenAddress"`
	TokenChain   *int    `json:"tokenChain"`
}
