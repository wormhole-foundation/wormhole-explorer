package chains

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
	rpcPool    map[sdk.ChainID]*pool.Pool
	p2pNetwork string
}

func fetchSeiDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*seiTx, error) {
	params := &cosmosTxSearchParams{Sequence: sequence, Timestamp: timestamp, SrcChannel: srcChannel, DstChannel: dstChannel}
	return fetchTxSearch[seiTx](ctx, baseUrl, params, seiTxSearchExtractor)
}

func (a *apiSei) fetchSeiTx(
	ctx context.Context,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {
	txHash = txHashLowerCaseWith0x(txHash)

	// get wormchain transaction
	wormchainTx, err := a.getWormchainTx(ctx, txHash)
	if err != nil {
		return nil, err
	}

	seiTx, err := fetchSeiDetail(ctx, baseUrl, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
	if err != nil {
		return nil, err
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

func (a *apiSei) getWormchainTx(ctx context.Context, txHash string) (*wormchainTx, error) {
	// Get the wormchain rpc pool
	wormchainPool, ok := a.rpcPool[sdk.ChainIDWormchain]
	if !ok {
		return nil, errors.New("wormchain rpc pool not found")
	}

	// Get the wormchain rpcs sorted by availability.
	wormchainRpcs := wormchainPool.GetItems()
	if len(wormchainRpcs) == 0 {
		return nil, errors.New("wormchain rpc pool is empty")
	}

	//var wormchainTx wormchainTx
	for _, rpc := range wormchainRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		// fetch wormchain transaction details
		wormchainTx, _ := fetchWormchainDetail(ctx, rpc.Id, txHash)
		if wormchainTx != nil {
			return wormchainTx, nil
		}
	}

	return nil, errors.New("failed to fetch wormchain transaction details")
}
