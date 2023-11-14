package chains

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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
	wormchainUrl         string
	wormchainRateLimiter *time.Ticker
	p2pNetwork           string
}

func fetchSeiDetail(ctx context.Context, baseUrl string, rateLimiter *time.Ticker, sequence, timestamp, srcChannel, dstChannel string) (*seiTx, error) {
	params := &cosmosTxSearchParams{Sequence: sequence, Timestamp: timestamp, SrcChannel: srcChannel, DstChannel: dstChannel}
	return fetchTxSearch[seiTx](ctx, baseUrl, rateLimiter, params, seiTxSearchExtractor)
}

func (a *apiSei) fetchSeiTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {
	txHash = txHashLowerCaseWith0x(txHash)
	wormchainTx, err := fetchWormchainDetail(ctx, a.wormchainUrl, a.wormchainRateLimiter, txHash)
	if err != nil {
		return nil, err
	}
	seiTx, err := fetchSeiDetail(ctx, baseUrl, rateLimiter, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
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
