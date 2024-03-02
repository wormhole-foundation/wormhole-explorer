package chains

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type seiTx struct {
	TxHash string
	Sender string
}

func seiTxSearchExtractor(tx *cosmosTxSearchResponse, logs []cosmosLogWrapperResponse) (*seiTx, error) {
	var sender string
	for _, l := range logs {
		for _, e := range l.Events {
			if e.Type == "message" {
				for _, attr := range e.Attributes {
					if attr.Key == "sender" {
						sender = attr.Value
					}
				}
				break
			}
		}
	}
	return &seiTx{TxHash: tx.Result.Txs[0].Hash, Sender: sender}, nil
}

type apiSei struct {
	p2pNetwork string
}

func fetchSeiDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*seiTx, error) {
	params := &cosmosTxSearchParams{Sequence: sequence, Timestamp: timestamp, SrcChannel: srcChannel, DstChannel: dstChannel}
	return fetchTxSearch[seiTx](ctx, baseUrl, params, seiTxSearchExtractor)
}

func (a *apiSei) fetchSeiTx(
	ctx context.Context,
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {
	txHash = txHashLowerCaseWith0x(txHash)

	// Get the wormchain rpc pool
	wormchainPool, ok := rpcPool[sdk.ChainIDWormchain]
	if !ok {
		return nil, errors.New("wormchain rpc pool not found")
	}

	// Get the wormchain rpcs sorted by availability.
	wormchainRpcs := wormchainPool.GetItems()
	if len(wormchainRpcs) == 0 {
		return nil, errors.New("wormchain rpc pool is empty")
	}

	var wormchainTx *wormchainTx
	var err error
	for _, rpc := range wormchainRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		// get wormchain transaction
		wormchainTx, err = fetchWormchainDetail(ctx, rpc.Id, txHash)
		if err != nil {
			continue
		}
		if wormchainTx != nil {
			break
		}
	}

	if wormchainTx == nil {
		return nil, errors.New("failed to fetch wormchain transaction details")
	}

	// Get the sei rpc pool
	seiPool, ok := rpcPool[sdk.ChainIDSei]
	if !ok {
		return nil, errors.New("sei rpc pool not found")
	}

	// Get the sei rpcs sorted by availability.
	seiRpcs := seiPool.GetItems()
	if len(seiRpcs) == 0 {
		return nil, errors.New("sei rpc pool is empty")
	}

	var seiTx *seiTx
	for _, rpc := range seiRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		// get sei transaction
		seiTx, err = fetchSeiDetail(ctx, rpc.Id, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
		if err != nil {
			logger.Error("failed to fetch sei transaction details", zap.Error(err))
			continue
		}
		if seiTx != nil {
			break
		}
	}

	if seiTx == nil {
		return nil, errors.New("failed to fetch sei transaction details")
	}

	return &TxDetail{
		NativeTxHash: txHash,
		From:         wormchainTx.receiver,
		Attribute: &AttributeTxDetail{
			Type: "wormchain-gateway",
			Value: &WorchainAttributeTxDetail{
				OriginChainID: sdk.ChainIDSei,
				OriginTxHash:  seiTx.TxHash,
				OriginAddress: seiTx.Sender,
			},
		},
	}, nil
}
