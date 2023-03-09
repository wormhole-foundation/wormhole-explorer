package watcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	MethodCompleteTransfer     = "completeTransfer"
	MethodWrapAndTransfer      = "wrapAndTransfer"
	MethodTransferTokens       = "transferTokens"
	MethodAttestToken          = "attestToken"
	MethodCompleteAndUnwrapETH = "completeAndUnwrapETH"
	MethodCreateWrapped        = "createWrapped"
	MethodUpdateWrapped        = "updateWrapped"
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

const (
	TxStatusFailedToProcess = "FAILED"
	TxStatusConfirmed       = "COMPLETED"
	TxStatusUnkonwn         = "UNKNOWN"
)

type EVMWatcher struct {
	client          *ankr.AnkrSDK
	chainID         vaa.ChainID
	blockchain      string
	contractAddress string
	repository      *storage.Repository
	logger          *zap.Logger
}

func NewEVMWatcher(client *ankr.AnkrSDK, chainID vaa.ChainID, blockchain, contractAddress string, repo *storage.Repository, logger *zap.Logger) *EVMWatcher {
	return &EVMWatcher{
		client:          client,
		chainID:         chainID,
		blockchain:      blockchain,
		contractAddress: contractAddress,
		repository:      repo,
		logger:          logger.With(zap.String("blockchain", blockchain)),
	}
}

func (w *EVMWatcher) Start(ctx context.Context) error {
	// get the current block for the chain.
	currentBlock, err := w.repository.GetCurrentBlock(ctx, w.blockchain)
	if err != nil {
		w.logger.Error("cannot get current block", zap.Error(err))
		return err
	}

	for {
		// get the latest block for the chain.
		stats, err := w.client.GetBlockchainStats(w.blockchain)
		if err != nil {
			w.logger.Error("cannot get blockchain stats", zap.Error(err))
		}
		if len(stats.Result.Stats) == 0 {
			return fmt.Errorf("no stats for blockchain %s", w.blockchain)
		}

		lastBlock := stats.Result.Stats[0].LatestBlockNumber
		if currentBlock < lastBlock {
			// process all the blocks between current and last block.
			w.processBlock(ctx, currentBlock, lastBlock)
		} else {
			time.Sleep(10 * time.Second)
		}
		currentBlock = lastBlock
	}

}

func (w *EVMWatcher) processBlock(ctx context.Context, currentBlock int64, lastBlock int64) {
	pageToken := ""
	hasPage := true

	for hasPage {
		// get the transactions
		request := ankr.NewTransactionsByAddressRequest(
			ankr.WithBlochchain(w.blockchain),
			ankr.WithContract(w.contractAddress),
			ankr.WithBlocks(currentBlock, lastBlock),
			ankr.WithPageToken(pageToken),
		)

		// get transaction data by address with pagination.
		r, err := w.client.GetTransactionsByAddress(*request)
		if err != nil {
			w.logger.Error("cannot get transactions by address", zap.Error(err))
			time.Sleep(10 * time.Second)
		}

		var lastBlockNumberHex string
		for _, tx := range r.Result.Transactions {
			w.logger.Debug("new tx", zap.String("tx", tx.Hash), zap.String("method", w.getMethodByInput(tx.Input)))
			switch w.getMethodByInput(tx.Input) {
			case MethodCompleteTransfer, MethodCompleteAndUnwrapETH, MethodCreateWrapped, MethodUpdateWrapped:
				// parse the VAA
				vaa, err := w.parseInput(tx.Input)
				if err != nil {
					w.logger.Error("cannot parse VAA", zap.Error(err), zap.String("tx", tx.Hash))
					continue
				}

				// create global transaction.
				updatedAt := time.Now()
				globalTx := storage.TransactionUpdate{
					ID: vaa.MessageID(),
					Destination: storage.DestinationTx{
						ChainID:     w.chainID,
						Status:      w.getTxStatus(tx.Status),
						Method:      w.getMethodByInput(tx.Input),
						TxHash:      tx.Hash,
						To:          tx.To,
						From:        tx.From,
						BlockNumber: tx.BlockNumber,
						Timestamp:   tx.Timestamp,
						UpdatedAt:   &updatedAt,
					},
				}
				err = w.repository.UpsertGlobalTransaction(ctx, globalTx)
				if err != nil {
					w.logger.Error("cannot save redeemed tx", zap.Error(err))
				}

				lastBlockNumberHex = tx.BlockNumber
			}
		}

		lastBlockNumber := strings.Replace(lastBlockNumberHex, "0x", "", -1)
		newBlockNumber, err := strconv.ParseInt(lastBlockNumber, 16, 64)
		if err != nil {
			w.logger.Error("error parsing block number", zap.Error(err), zap.String("blockNumber", lastBlockNumber))
			continue
		}

		watcherBlock := storage.WatcherBlock{
			ID:          w.blockchain,
			BlockNumber: newBlockNumber,
		}
		w.repository.UpdateWatcherBlock(ctx, watcherBlock)

		pageToken := r.Result.NextPageToken
		if pageToken == "" {
			hasPage = false
		}
	}
}

func (w *EVMWatcher) Close() {
}

// get transaction status
func (w *EVMWatcher) getTxStatus(status string) string {
	switch status {
	case TxStatusSuccess:
		return TxStatusConfirmed
	case TxStatusFailReverted:
		return TxStatusFailedToProcess
	default:
		return fmt.Sprintf("%s: %s", TxStatusUnkonwn, status)
	}
}

// get executed method by input
// completeTransfer, completeAndUnwrapETH, createWrapped receive a VAA as input
func (w *EVMWatcher) getMethodByInput(input string) string {
	method := input[0:10]
	switch method {
	case MethodIDCompleteTransfer:
		return MethodCompleteTransfer
	case MethodIDWrapAndTransfer:
		return MethodWrapAndTransfer
	case MethodIDTransferTokens:
		return MethodTransferTokens
	case MethodIDAttestToken:
		return MethodAttestToken
	case MethodIDCompleteAndUnwrapETH:
		return MethodCompleteAndUnwrapETH
	case MethodIDCreateWrapped:
		return MethodCreateWrapped
	case MethodIDUpdateWrapped:
		return MethodUpdateWrapped
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
