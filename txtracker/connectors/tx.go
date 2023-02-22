package connectors

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const (
	TokenBridgeBsc      = "0xb6f6d86a8f9879a9c87f643768d9efc38c1da6e7"
	TokenBridgeEthereum = "0x3ee18b2214aff97000d974cf647e7c347e8fa585"
	TokenBridgePolygon  = "0x5a58505a96d1dbf8df91cb21b54419fc36e93fde"
)

type TxData struct {
	Source      string
	Destination string
	Amount      *big.Int
	Decimals    uint8
	Timestamp   time.Time
}

func FetchTx(
	ctx context.Context,
	cfg *config.Settings,
	chainId vaa.ChainID,
	txHash string,
) (*TxData, error) {

	// decide which RPC/API service to use based on chain ID
	var fetchFunc func(context.Context, *config.Settings, string) (*TxData, error)
	switch chainId {
	case vaa.ChainIDEthereum:
		fetchFunc = ankrFetchEthTx
	case vaa.ChainIDBSC:
		fetchFunc = ankrFetchBscTx
	case vaa.ChainIDPolygon:
		fetchFunc = ankrFetchPolygonTx
	case vaa.ChainIDSolana:
		fetchFunc = fetchSolanaTx
	case vaa.ChainIDTerra:
		fetchFunc = fetchTerraTx
	}
	if fetchFunc == nil {
		return nil, fmt.Errorf("chain ID not supported: %v", chainId)
	}

	// get transaction details from the service
	txData, err := fetchFunc(ctx, cfg, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tx information: %w", err)
	}

	return txData, nil
}
