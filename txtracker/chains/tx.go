package chains

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const (
	TokenBridgeBsc      = "0xb6f6d86a8f9879a9c87f643768d9efc38c1da6e7"
	TokenBridgeEthereum = "0x3ee18b2214aff97000d974cf647e7c347e8fa585"
	TokenBridgePolygon  = "0x5a58505a96d1dbf8df91cb21b54419fc36e93fde"
)

const requestTimeout = 10 * time.Second

const topicAttestToken = "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2"

var (
	ErrChainNotSupported = errors.New("chain id not supported")
)

type TxDetail struct {
	Source      string
	Destination string
	Timestamp   time.Time
}

var tickers = struct {
	ankr        *time.Ticker
	blockdaemon *time.Ticker
	solana      *time.Ticker
	terra       *time.Ticker
}{}

func init() {

	tickers.ankr = time.NewTicker(2 * time.Second)
	tickers.blockdaemon = time.NewTicker(5 * time.Second)
	tickers.terra = time.NewTicker(5 * time.Second)

	// the Solana adapter sends 2 requests per txHash
	tickers.solana = time.NewTicker(10 * time.Second)
}

func FetchTx(
	ctx context.Context,
	cfg *config.Settings,
	chainId vaa.ChainID,
	txHash string,
) (*TxDetail, error) {

	// decide which RPC/API service to use based on chain ID
	var fetchFunc func(context.Context, *config.Settings, string) (*TxDetail, error)
	var rateLimiter time.Ticker
	switch chainId {
	case vaa.ChainIDEthereum:
		fetchFunc = ankrFetchEthTx
		rateLimiter = *tickers.ankr
	case vaa.ChainIDBSC:
		fetchFunc = ankrFetchBscTx
		rateLimiter = *tickers.ankr
	case vaa.ChainIDPolygon:
		fetchFunc = ankrFetchPolygonTx
		rateLimiter = *tickers.ankr
	//case vaa.ChainIDSolana:
	//	fetchFunc = fetchSolanaTx
	//	rateLimiter = *tickers.solana
	case vaa.ChainIDTerra:
		fetchFunc = fetchTerraTx
		rateLimiter = *tickers.terra
	default:
		return nil, ErrChainNotSupported
	}
	if fetchFunc == nil {
		return nil, fmt.Errorf("chain ID not supported: %v", chainId)
	}

	// wait for rate limit - fail fast if context was cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-rateLimiter.C:
	}

	// get transaction details from the RPC/API service
	subContext, cancelFunc := context.WithTimeout(ctx, requestTimeout)
	defer cancelFunc()
	txDetail, err := fetchFunc(subContext, cfg, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tx information: %w", err)
	}

	return txDetail, nil
}
