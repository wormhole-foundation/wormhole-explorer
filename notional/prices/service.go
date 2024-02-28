package prices

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

type Price struct {
	CoingeckoID string    `json:"coingeckoId"`
	Symbol      string    `json:"symbol"`
	Price       string    `json:"price"`
	Datetime    time.Time `json:"dateTime"`
}

// PriceService provides an interface to interact with prices.
type PriceService struct {
	priceRepository *PriceRepository
	tokenProvider   *domain.TokenProvider
	notionalCache   wormscanNotionalCache.NotionalLocalCacheReadable
	logger          *zap.Logger
}

// NewPriceService creates a new price service.
func NewPriceService(priceRepository *PriceRepository,
	tokenProvider *domain.TokenProvider,
	notionalCache wormscanNotionalCache.NotionalLocalCacheReadable,
	logger *zap.Logger) *PriceService {
	return &PriceService{
		priceRepository: priceRepository,
		tokenProvider:   tokenProvider,
		notionalCache:   notionalCache,
		logger:          logger.With(zap.String("module", "priceService")),
	}
}

// GetPrice returns the price of a token at a given datetime.
func (s *PriceService) GetPrice(ctx context.Context, tokenChainID sdk.ChainID, tokenAddress string, datetime time.Time) (*Price, error) {
	log := s.logger.With(zap.Uint16("chainID", uint16(tokenChainID)), zap.String("tokenAddress", tokenAddress))
	token, found := s.tokenProvider.GetTokenByAddress(tokenChainID, tokenAddress)
	if !found {
		log.Warn("Token not found")
		return nil, ErrTokenNotFound
	}

	v, err := s.GetPriceBySymbol(ctx, token, datetime, log)
	if err != nil {
		return nil, err
	}

	return &Price{
		CoingeckoID: token.CoingeckoID,
		Symbol:      token.Symbol.String(),
		Price:       v.price.String(),
		Datetime:    v.datetime,
	}, nil
}

func (s *PriceService) GetPriceByCoingeckoID(ctx context.Context, coingeckoID string, datetime time.Time) (*Price, error) {
	log := s.logger.With(zap.String("coingeckoId", coingeckoID))
	token, found := s.tokenProvider.GetTokenByCoingeckoID(coingeckoID)
	if !found {
		log.Warn("Token not found")
		return nil, ErrTokenNotFound
	}
	v, err := s.GetPriceBySymbol(ctx, token, datetime, log)
	if err != nil {
		return nil, err
	}

	return &Price{
		CoingeckoID: coingeckoID,
		Symbol:      token.Symbol.String(),
		Price:       v.price.String(),
		Datetime:    v.datetime,
	}, nil
}

type priceSymbol struct {
	price    decimal.Decimal
	datetime time.Time
}

func (s *PriceService) GetPriceBySymbol(ctx context.Context, token *domain.TokenMetadata, datetime time.Time, log *zap.Logger) (*priceSymbol, error) {
	if token.CoingeckoID == "" {
		log.Warn("CoingeckoID not found")
		return nil, ErrTokenNotFound
	}

	dayDatetime := datetime.Truncate(24 * time.Hour)
	cachePrice, err := s.notionalCache.Get(token.GetTokenID())
	if err != nil {
		if err == wormscanNotionalCache.ErrNotFound {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	diffCachePrice := cachePrice.UpdatedAt.Sub(datetime).Abs()
	diffDayPrice := dayDatetime.Sub(datetime).Abs()
	var price decimal.Decimal
	var priceDatetime time.Time
	if diffCachePrice > diffDayPrice {
		p, err := s.priceRepository.Find(ctx, token.CoingeckoID, dayDatetime)
		if err != nil {
			if err == ErrPriceNotFound {
				return nil, ErrTokenNotFound
			}
			log.Error("Failed to find price", zap.Error(err))
			return nil, err
		}
		price, err = decimal.NewFromString(p.Price)
		if err != nil {
			log.Error("Failed to parse price", zap.Error(err))
			return nil, err
		}
		priceDatetime = p.Datetime
	} else {
		price = cachePrice.NotionalUsd
		priceDatetime = cachePrice.UpdatedAt
	}
	return &priceSymbol{price: price, datetime: priceDatetime}, nil
}
