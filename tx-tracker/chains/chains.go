package chains

import (
	"context"
	"errors"
	"fmt"
	notional "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var (
	ErrChainNotSupported   = errors.New("chain id not supported")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type TxDetail struct {
	// From is the address that signed the transaction, encoded in the chain's native format.
	From string
	// NativeTxHash contains the transaction hash, encoded in the chain's native format.
	NativeTxHash string
	// Attribute contains the specific information of the transaction.
	Attribute *AttributeTxDetail
	// FeeDetail contains the fee of the transactions.
	FeeDetail *FeeDetail
}

type FeeDetail struct {
	Fee    string            `bson:"fee" json:"fee"`
	RawFee map[string]string `bson:"rawFee" json:"rawFee"`
	FeeUSD float64           `bson:"feeUSD" json:"feeUSD"`
}

type AttributeTxDetail struct {
	Type  string
	Value any
}

func FetchTx(ctx context.Context, rpcPool map[sdk.ChainID]*pool.Pool, wormchainRpcPool map[sdk.ChainID]*pool.Pool, chainId sdk.ChainID, txHash string, timestamp *time.Time, p2pNetwork string, m metrics.Metrics, logger *zap.Logger, notionalCache *notional.NotionalCache) (*TxDetail, error) {
	// Decide which RPC/API service to use based on chain ID
	var fetchFunc func(ctx context.Context, pool *pool.Pool, txHash string, metrics metrics.Metrics, logger *zap.Logger) (*TxDetail, error)
	switch chainId {
	case sdk.ChainIDSolana:
		apiSolana := &apiSolana{
			timestamp:     timestamp,
			notionalCache: notionalCache,
		}
		fetchFunc = apiSolana.FetchSolanaTx
	case sdk.ChainIDAlgorand:
		fetchFunc = FetchAlgorandTx
	case sdk.ChainIDAptos:
		fetchFunc = FetchAptosTx
	case sdk.ChainIDSui:
		fetchFunc = FetchSuiTx
	case sdk.ChainIDInjective,
		sdk.ChainIDTerra,
		sdk.ChainIDTerra2,
		sdk.ChainIDXpla:
		apiCosmos := &apiCosmos{
			chainId: chainId,
		}
		fetchFunc = apiCosmos.FetchCosmosTx
	case sdk.ChainIDAcala,
		sdk.ChainIDArbitrum,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDAvalanche,
		sdk.ChainIDBase,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDBSC,
		sdk.ChainIDCelo,
		sdk.ChainIDEthereum,
		sdk.ChainIDSepolia,
		sdk.ChainIDFantom,
		sdk.ChainIDKarura,
		sdk.ChainIDKlaytn,
		sdk.ChainIDMoonbeam,
		sdk.ChainIDOasis,
		sdk.ChainIDOptimism,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDPolygon,
		sdk.ChainIDScroll,
		sdk.ChainIDBlast,
		sdk.ChainIDXLayer,
		sdk.ChainIDMantle,
		sdk.ChainIDPolygonSepolia: // polygon amoy
		apiEvm := &apiEvm{
			chainId:       chainId,
			timestamp:     timestamp,
			notionalCache: notionalCache,
		}
		fetchFunc = apiEvm.FetchEvmTx
	case sdk.ChainIDWormchain:
		apiWormchain := &apiWormchain{
			p2pNetwork:    p2pNetwork,
			evmosPool:     wormchainRpcPool[sdk.ChainIDEvmos],
			kujiraPool:    wormchainRpcPool[sdk.ChainIDKujira],
			osmosisPool:   wormchainRpcPool[sdk.ChainIDOsmosis],
			injectivePool: wormchainRpcPool[sdk.ChainIDInjective],
		}
		fetchFunc = apiWormchain.FetchWormchainTx
	case sdk.ChainIDSei:
		apiSei := &apiSei{
			p2pNetwork:    p2pNetwork,
			wormchainPool: rpcPool[sdk.ChainIDWormchain],
		}
		fetchFunc = apiSei.FetchSeiTx
	default:
		return nil, ErrChainNotSupported
	}

	pool, ok := rpcPool[chainId]
	if !ok {
		return nil, fmt.Errorf("not found rpc pool for chain %s", chainId.String())
	}

	txDetail, err := fetchFunc(ctx, pool, txHash, m, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tx information: %w", err)
	}

	return txDetail, nil
}
