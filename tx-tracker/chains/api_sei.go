package chains

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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
	p2pNetwork    string
	wormchainPool *pool.Pool
}

func fetchSeiDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*seiTx, error) {
	params := &cosmosTxSearchParams{Sequence: sequence, Timestamp: timestamp, SrcChannel: srcChannel, DstChannel: dstChannel}
	return fetchTxSearch[seiTx](ctx, baseUrl, params, seiTxSearchExtractor)
}

func (a *apiSei) FetchSeiTx(
	ctx context.Context,
	pool *pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {
	txHash = txHashLowerCaseWith0x(txHash)

	// Get the wormchain rpcs sorted by availability.
	wormchainRpcs := a.wormchainPool.GetItems()
	if len(wormchainRpcs) == 0 {
		return nil, errors.New("wormchain rpc pool is empty")
	}

	// Fetch the wormchain transaction
	var wormchainTx *wormchainTx
	var err error
	for _, rpc := range wormchainRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		wormchainTx, err = fetchWormchainDetail(ctx, rpc.Id, txHash)
		if err != nil {
			continue
		}
		break
	}

	// If the transaction is not found, return an error
	if err != nil {
		return nil, err
	}
	if wormchainTx == nil {
		return nil, ErrTransactionNotFound
	}

	// Get the sei rpcs sorted by availability.
	seiRpcs := pool.GetItems()
	if len(seiRpcs) == 0 {
		return nil, errors.New("sei rpc pool is empty")
	}

	// Fetch the sei transaction
	var seiTx *seiTx
	for _, rpc := range seiRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		seiTx, err = fetchSeiDetail(ctx, rpc.Id, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
		if err != nil {
			continue
		}
		break
	}

	// If the transaction is not found, return an error
	if err != nil {
		return nil, err
	}
	if seiTx == nil {
		return nil, ErrTransactionNotFound
	}

	return &TxDetail{
		NativeTxHash: txHash,
		From:         wormchainTx.receiver,
		Attribute: &AttributeTxDetail{
			Type: "wormchain-gateway",
			Value: &WorchainAttributeTxDetail{
				OriginChainID: vaa.ChainIDSei,
				OriginTxHash:  seiTx.TxHash,
				OriginAddress: seiTx.Sender,
			},
		},
	}, nil
}
