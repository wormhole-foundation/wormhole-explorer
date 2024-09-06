package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const nttTopHolders = "wormscan:ntt-top-holders"

type HolderRepositoryReadable struct {
	cache cache.Cache
	log   *zap.Logger
}

type HolderRepository struct {
	client        *resty.Client
	arkhamUrl     string
	artkhamApiKey string
	solanaUrl     string
	cache         cache.Cache
	tokenProvider *domain.TokenProvider
	notionalCache notional.NotionalLocalCacheReadable
	log           *zap.Logger
}

func NewHolderRepository(client *resty.Client, arkhamUrl, arkhamApiKey, solanaUrl string, cache cache.Cache,
	tokenProvider *domain.TokenProvider,
	notionalCache notional.NotionalLocalCacheReadable,
	log *zap.Logger) *HolderRepository {
	return &HolderRepository{
		client:        resty.New(),
		arkhamUrl:     arkhamUrl,
		artkhamApiKey: arkhamApiKey,
		solanaUrl:     solanaUrl,
		cache:         cache,
		notionalCache: notionalCache,
		tokenProvider: tokenProvider,
		log:           log,
	}
}

func (r *HolderRepository) LoadNativeTokenTransferTopHolder(ctx context.Context, symbol string, expiration time.Duration) error {
	holders, err := r.getNativeTokenTransferTopHolder(ctx, symbol)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s:%s", nttTopHolders, symbol)
	cr := cachedResult[[]NativeTokenTransferTopHolder]{Timestamp: time.Now(), Result: holders}
	return r.cache.Set(ctx, key, cr, expiration)
}

func (r *HolderRepository) GetNativeTokenTransferTopHolder(ctx context.Context, symbol string) ([]NativeTokenTransferTopHolder, error) {
	key := fmt.Sprintf("%s:%s", nttTopHolders, symbol)
	result, err := r.cache.Get(ctx, key)
	if err != nil {
		return r.getNativeTokenTransferTopHolder(ctx, symbol)
	}
	var cached cachedResult[[]NativeTokenTransferTopHolder]
	err = json.Unmarshal([]byte(result), &cached)
	if err != nil {
		return nil, err
	}
	return cached.Result, nil
}

func (r *HolderRepository) getNativeTokenTransferTopHolder(ctx context.Context, symbol string) ([]NativeTokenTransferTopHolder, error) {
	holderEVM, err := r.getHolderByWormholeTokenForEVM(ctx)
	if err != nil {
		return nil, err
	}

	holderSolana, err := r.getHolderByWormholeTokenForSolana(ctx)
	if err != nil {
		return nil, err
	}

	// Merge holders from different chains
	var holders []NativeTokenTransferTopHolder
	for _, holder := range holderEVM.AddressTopHolders.Ethereum {
		holders = append(holders, NativeTokenTransferTopHolder{
			Address: holder.Address.Address,
			ChainID: sdk.ChainIDEthereum,
			Volume:  decimal.NewFromFloat(holder.Usd),
		})
	}

	for _, holder := range holderEVM.AddressTopHolders.ArbitrumOne {
		holders = append(holders, NativeTokenTransferTopHolder{
			Address: holder.Address.Address,
			ChainID: sdk.ChainIDArbitrum,
			Volume:  decimal.NewFromFloat(holder.Usd),
		})
	}

	for _, holder := range holderEVM.AddressTopHolders.Base {
		holders = append(holders, NativeTokenTransferTopHolder{
			Address: holder.Address.Address,
			ChainID: sdk.ChainIDBase,
			Volume:  decimal.NewFromFloat(holder.Usd),
		})
	}

	// Get token metadata price for solana
	tokens, found := r.tokenProvider.GetTokensBySymbol(symbol)
	if !found {
		return nil, fmt.Errorf("no token found for symbol %s", symbol)

	}
	var token *domain.TokenMetadata
	for _, t := range tokens {
		if t.TokenChain == sdk.ChainIDSolana {
			token = t
			break
		}
	}
	if token == nil {
		return nil, fmt.Errorf("no token found for symbol %s", symbol)
	}

	tokenPrice, err := r.notionalCache.Get(token.GetTokenID())
	if err != nil {
		return nil, err
	}

	// Add holders from solana with calculated price
	for _, holder := range holderSolana.Result.Value {
		holders = append(holders, NativeTokenTransferTopHolder{
			Address: holder.Address,
			ChainID: sdk.ChainIDSolana,
			Volume:  decimal.NewFromFloat(holder.UIAmount).Mul(tokenPrice.NotionalUsd),
		})
	}

	// Sort holders by price in descending order
	sort.Slice(holders, func(i, j int) bool {
		return holders[i].Volume.Compare(holders[j].Volume) > 0
	})

	// Return top 10 holders
	if len(holders) > 10 {
		holders = holders[:10]
	}

	return holders, nil
}

func (r *HolderRepository) getHolderByWormholeTokenForEVM(ctx context.Context) (*topHoldersEVMResponse, error) {
	url := fmt.Sprintf("%s/token/holders/wormhole", r.arkhamUrl)
	resp, err := r.client.R().
		SetContext(ctx).
		SetHeader("API-Key", r.artkhamApiKey).
		SetResult(&topHoldersEVMResponse{}).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*topHoldersEVMResponse)
	if result == nil {
		return nil, fmt.Errorf("empty response")
	}
	return result, nil
}

func (r *HolderRepository) getHolderByWormholeTokenForSolana(ctx context.Context) (*topHoldersSolanaResponse, error) {
	resp, err := r.client.R().
		SetContext(ctx).
		SetBody(`{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "getTokenLargestAccounts",
			"params": [ "85VBFQZC9TZkfaptBWjvUw7YbZjy52A6mjtPGjstQAmQ"]
		}`).
		SetResult(&topHoldersSolanaResponse{}).
		Post(r.solanaUrl)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*topHoldersSolanaResponse)
	if result == nil {
		return nil, fmt.Errorf("empty response")
	}
	return result, nil
}

type topHoldersEVMResponse struct {
	AddressTopHolders struct {
		ArbitrumOne []struct {
			Address struct {
				Address string `json:"address"`
			} `json:"address"`
			Usd float64 `json:"usd"`
		} `json:"arbitrum_one"`
		Base []struct {
			Address struct {
				Address string `json:"address"`
			} `json:"address"`
			Usd float64 `json:"usd"`
		} `json:"base"`
		Ethereum []struct {
			Address struct {
				Address string `json:"address"`
			} `json:"address"`
			Usd float64 `json:"usd"`
		} `json:"ethereum"`
	} `json:"addressTopHolders"`
}

type topHoldersSolanaResponse struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value []struct {
			Address        string  `json:"address"`
			Amount         string  `json:"amount"`
			Decimals       int     `json:"decimals"`
			UIAmount       float64 `json:"uiAmount"`
			UIAmountString string  `json:"uiAmountString"`
		} `json:"value"`
	} `json:"result"`
}

func NewHolderRepositoryReadable(cache cache.Cache, log *zap.Logger) *HolderRepositoryReadable {
	return &HolderRepositoryReadable{
		cache: cache,
		log:   log,
	}
}

func (r *HolderRepositoryReadable) GetNativeTokenTransferTopHolder(ctx context.Context, symbol string) ([]NativeTokenTransferTopHolder, error) {
	key := fmt.Sprintf("%s:%s", nttTopHolders, symbol)
	result, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var cached cachedResult[[]NativeTokenTransferTopHolder]
	err = json.Unmarshal([]byte(result), &cached)
	if err != nil {
		return nil, err
	}
	return cached.Result, nil
}
