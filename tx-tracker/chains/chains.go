package chains

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
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
}

type AttributeTxDetail struct {
	Type  string
	Value any
}

func FetchTx(
	ctx context.Context,
	rpcPool map[sdk.ChainID]*pool.Pool,
	chainId sdk.ChainID,
	txHash string,
	timestamp *time.Time,
	p2pNetwork string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Decide which RPC/API service to use based on chain ID
	var fetchFunc func(ctx context.Context, chainID sdk.ChainID, rpcPool map[sdk.ChainID]*pool.Pool, txHash string, logger *zap.Logger) (*TxDetail, error)
	switch chainId {
	case sdk.ChainIDSolana:
		apiSolana := &apiSolana{
			timestamp: timestamp,
		}
		fetchFunc = apiSolana.fetchSolanaTx
	case sdk.ChainIDAlgorand:
		fetchFunc = fetchAlgorandTx
	case sdk.ChainIDAptos:
		fetchFunc = fetchAptosTx
	case sdk.ChainIDSui:
		fetchFunc = fetchSuiTx
	case sdk.ChainIDInjective,
		sdk.ChainIDTerra,
		sdk.ChainIDTerra2,
		sdk.ChainIDXpla:
		fetchFunc = fetchCosmosTx
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
		sdk.ChainIDPolygon:
		fetchFunc = fetchEvmTx
	case sdk.ChainIDWormchain:
		apiWormchain := &apiWormchain{
			p2pNetwork: p2pNetwork,
		}
		fetchFunc = apiWormchain.fetchWormchainTx
	case sdk.ChainIDSei:
		apiSei := &apiSei{
			p2pNetwork: p2pNetwork,
		}
		fetchFunc = apiSei.fetchSeiTx

	default:
		return nil, ErrChainNotSupported
	}

	// pool, ok := rpcPool[chainId]
	// if !ok {
	// 	return nil, fmt.Errorf("not found rpc pool for chain %s", chainId.String())
	// }

	// // get rpc sorted by score and priority.
	// rpcs := pool.GetItems()
	// if len(rpcs) == 0 {
	// 	logger.Error("not found rpc pool", zap.String("chainId", chainId.String()))
	// 	return nil, ErrChainNotSupported
	// }

	//for _, rpc := range rpcs {
	// Fetch transaction details from the RPC/API service
	//rpc.Wait(ctx)
	TxDetail, err := fetchFunc(ctx, chainId, rpcPool, txHash, logger)
	if err == nil {
		logger.Debug("Fetched transaction details",
			zap.String("txHash", txHash),
			zap.String("chainId", chainId.String()),
			zap.String("from", TxDetail.From))
		return TxDetail, nil
	}
	//}

	return nil, errors.New("failed to fetch transaction details")
}

// getRpcPool returns the rpc pool for the given chain ID.
func getRpcPool(rpcPool map[sdk.ChainID]*pool.Pool, chainId sdk.ChainID) ([]pool.Item, error) {
	pool, ok := rpcPool[chainId]
	if !ok {
		return nil, fmt.Errorf("not found rpc pool for chain %s", chainId.String())
	}

	// get rpc sorted by score and priority.
	rpcs := pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	return rpcs, nil
}
