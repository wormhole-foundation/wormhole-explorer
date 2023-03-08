package watcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	MethodIDCompleteTransfer     = "0xc6878519"
	MethodIDWrapAndTransfer      = "0x9981509f"
	MethodIDTransferTokens       = "0x0f5287b0"
	MethodIDAttestToken          = "0xc48fa115"
	MethodIDCompleteAndUnwrapETH = "0xff200cde"
	MethodIDCreateWrapped        = "0xe8059810"
	MethodIDUpdateWrapped        = "0xf768441f"
)

const (
	TxStatusSuccess      = "0x1"
	TxStatusFailReverted = "0x0"
)

type EVMWatcher struct {
	client          *ankr.AnkrSDK
	blockchain      string
	contractAddress string
	repository      *storage.Repository
	logger          *zap.Logger
}

func NewEVMWatcher(client *ankr.AnkrSDK, blockchain, contractAddress string, repo *storage.Repository, logger *zap.Logger) *EVMWatcher {
	return &EVMWatcher{
		client:          client,
		blockchain:      blockchain,
		contractAddress: contractAddress,
		repository:      repo,
		logger:          logger.With(zap.String("blockchain", blockchain)),
	}
}

func (w *EVMWatcher) Start(ctx context.Context) error {

	var lastBlock int64
	stats, err := w.client.GetBlockchainStats(w.blockchain)
	if err != nil {
		w.logger.Error("cannot get blockchain stats", zap.Error(err))
	}
	if len(stats.Result.Stats) == 0 {
		return fmt.Errorf("no stats for blockchain %s", w.blockchain)
	}

	lastBlock = stats.Result.Stats[0].LatestBlockNumber

	w.logger.Info("Starting", zap.Int64("lastBlock", lastBlock))

	for {
		// get the last block
		stats, err := w.client.GetBlockchainStats(w.blockchain)
		if err != nil {
			w.logger.Error("cannot get blockchain stats", zap.Error(err))
		}

		if len(stats.Result.Stats) == 0 {
			w.logger.Warn("no stats for blockchain", zap.String("blockchain", w.blockchain))
			time.Sleep(10 * time.Second) // cool off
			continue
		}

		if stats.Result.Stats[0].LatestBlockNumber > lastBlock {

			w.logger.Info("new block", zap.Int64("lastBlock", lastBlock), zap.Int64("latestBlock", stats.Result.Stats[0].LatestBlockNumber))

			// get the transactions
			request := ankr.NewTransactionsByAddressRequest(
				ankr.WithBlochchain(w.blockchain),
				ankr.WithContract(w.contractAddress),
				ankr.WithBlocks(lastBlock, stats.Result.Stats[0].LatestBlockNumber),
			)

			r, err := w.client.GetTransactionsByAddress(*request)
			if err != nil {
				w.logger.Error("cannot get transactions by address", zap.Error(err))
			}

			for _, tx := range r.Result.Transactions {
				w.logger.Info("new tx", zap.String("tx", tx.Hash), zap.String("method", w.getMethodByInput(tx.Input)))
				switch w.getMethodByInput(tx.Input) {
				case "completeTransfer", "completeAndUnwrapETH", "createWrapped", "updateWrapped":
					vaa, err := w.parseInput(tx.Input)
					if err != nil {
						w.logger.Error("cannot parse VAA", zap.Error(err))
					} else {
						redeemed := storage.RedeemedUpdate{
							ID:           vaa.MessageID(),
							Chain:        request.RquestParams.Blockchain, //redeemed chain
							TxHash:       tx.Hash,
							Method:       w.getMethodByInput(tx.Input),
							EmitterChain: vaa.EmitterChain,
							EmitterAddr:  vaa.EmitterAddress.String(),
							Sequence:     fmt.Sprintf("%d", vaa.Sequence),
							To:           tx.To,
							From:         tx.From,
							BlockNumber:  tx.BlockNumber,
							Status:       w.getTxStatus(tx.Status),
							VaaTimestamp: &vaa.Timestamp,
						}

						err = w.repository.UpsertRedeemed(ctx, redeemed)
						if err != nil {
							w.logger.Error("cannot save redeemed tx", zap.Error(err))
						}
					}

				}

			}

			lastBlock = stats.Result.Stats[0].LatestBlockNumber

		} else {
			w.logger.Info("no new blocks", zap.Int64("lastBlock", lastBlock))
		}

		time.Sleep(12 * time.Second)
	}

}

func (w *EVMWatcher) Close() {
}

// get transaction status
func (w *EVMWatcher) getTxStatus(status string) string {
	switch status {
	case TxStatusSuccess:
		return "success"
	case TxStatusFailReverted:
		return "fail (reverted)"
	default:
		return fmt.Sprintf("unknown (%s)", status)
	}
}

// get executed method by input
// completeTransfer, completeAndUnwrapETH, createWrapped receive a VAA as input
func (w *EVMWatcher) getMethodByInput(input string) string {
	method := input[0:10]
	switch method {
	case MethodIDCompleteTransfer:
		return "completeTransfer"
	case MethodIDWrapAndTransfer:
		return "wrapAndTransfer"
	case MethodIDTransferTokens:
		return "transferTokens"
	case MethodIDAttestToken:
		return "attestToken"
	case MethodIDCompleteAndUnwrapETH:
		return "completeAndUnwrapETH"
	case MethodIDCreateWrapped:
		return "createWrapped"
	case MethodIDUpdateWrapped:
		return "updateWrapped"
	default:
		return fmt.Sprintf("unknown (%s)", method)

	}
}

// get the input and extract the method signature and VAA
func (w *EVMWatcher) parseInput(input string) (*vaa.VAA, error) {
	// remove the first 64 characters plus 0x
	input = input[138:]
	vaaBytes, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}

	vaa, err := vaa.Unmarshal(vaaBytes)
	if err != nil {
		return nil, err
	}

	return vaa, nil
}
