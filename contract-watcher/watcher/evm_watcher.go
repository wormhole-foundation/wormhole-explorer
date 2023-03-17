package watcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	MethodUnkown               = "unknown"
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
	chainID         vaa.ChainID
	blockchain      string
	contractAddress string
	sizeBlocks      uint8
	waitSeconds     uint16
	initialBlock    int64
	repository      *storage.Repository
	logger          *zap.Logger
	close           chan bool
	wg              sync.WaitGroup
}
type EVMParams struct {
	ChainID         vaa.ChainID
	Blockchain      string
	ContractAddress string
	SizeBlocks      uint8
	WaitSeconds     uint16
	InitialBlock    int64
}

func NewEVMWatcher(client *ankr.AnkrSDK, repo *storage.Repository, params EVMParams, logger *zap.Logger) *EVMWatcher {
	return &EVMWatcher{
		client:          client,
		chainID:         params.ChainID,
		blockchain:      params.Blockchain,
		contractAddress: params.ContractAddress,
		sizeBlocks:      params.SizeBlocks,
		waitSeconds:     params.WaitSeconds,
		initialBlock:    params.InitialBlock,
		repository:      repo,
		logger:          logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(params.ChainID))),
	}
}

func (w *EVMWatcher) Start(ctx context.Context) error {
	// get the current block for the chain.
	currentBlock, err := w.repository.GetCurrentBlock(ctx, w.blockchain, w.initialBlock)
	if err != nil {
		w.logger.Error("cannot get current block", zap.Error(err))
		return err
	}

	w.wg.Add(1)
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("clossing watcher by context")
			w.wg.Done()
			return nil
		case <-w.close:
			w.logger.Info("clossing watcher")
			w.wg.Done()
			return nil
		default:
			// get the latest block for the chain.
			stats, err := w.client.GetBlockchainStats(ctx, w.blockchain)
			if err != nil {
				w.logger.Error("cannot get blockchain stats", zap.Error(err))
			}
			if len(stats.Result.Stats) == 0 {
				return fmt.Errorf("no stats for blockchain %s", w.blockchain)
			}

			maxBlocks := int64(w.sizeBlocks)
			lastBlock := stats.Result.Stats[0].LatestBlockNumber
			if currentBlock < lastBlock {
				totalBlocks := (lastBlock-currentBlock)/maxBlocks + 1
				for i := 0; i < int(totalBlocks); i++ {
					fromBlock := currentBlock + int64(i)*maxBlocks
					toBlock := fromBlock + maxBlocks - 1
					if toBlock > lastBlock {
						toBlock = lastBlock
					}
					w.logger.Info("processing blocks", zap.Int64("from", fromBlock), zap.Int64("to", toBlock))
					w.processBlock(ctx, fromBlock, toBlock)
					w.logger.Info("blocks processed", zap.Int64("from", fromBlock), zap.Int64("to", toBlock))
				}
				// process all the blocks between current and last block.
			} else {
				w.logger.Info("waiting for new blocks")
				select {
				case <-ctx.Done():
					w.wg.Done()
					return nil
				case <-time.After(time.Duration(w.waitSeconds) * time.Second):
				}
			}
			currentBlock = lastBlock
		}
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
		r, err := w.client.GetTransactionsByAddress(ctx, *request)
		if err != nil {
			w.logger.Error("cannot get transactions by address", zap.Error(err))
			time.Sleep(10 * time.Second)
		}

		var lastBlockNumberHex string
		for _, tx := range r.Result.Transactions {
			w.logger.Debug("new tx", zap.String("tx", tx.Hash), zap.String("method", w.getMethodByInput(tx.Input)))
			method := w.getMethodByInput(tx.Input)
			switch method {
			case MethodCompleteTransfer, MethodCompleteAndUnwrapETH, MethodCreateWrapped, MethodUpdateWrapped:
				// parse the VAA
				vaa, err := w.parseInput(tx.Input)
				if err != nil {
					w.logger.Error("cannot parse VAA", zap.Error(err), zap.String("tx", tx.Hash))
					continue
				}

				// get the timestamp.
				unixTime, err := strconv.ParseInt(remove0x(tx.Timestamp), 16, 64)
				var timestamp *time.Time
				if err != nil {
					w.logger.Error("cannot convert to timestamp", zap.Error(err), zap.String("tx", tx.Hash))
				} else {
					tm := time.Unix(unixTime, 0)
					timestamp = &tm
				}

				// create global transaction.
				updatedAt := time.Now()
				globalTx := storage.TransactionUpdate{
					ID: vaa.MessageID(),
					Destination: storage.DestinationTx{
						ChainID:     w.chainID,
						Status:      w.getTxStatus(tx.Status),
						Method:      w.getMethodByInput(tx.Input),
						TxHash:      remove0x(tx.Hash),
						To:          tx.To,
						From:        tx.From,
						BlockNumber: tx.BlockNumber,
						Timestamp:   timestamp,
						UpdatedAt:   &updatedAt,
					},
				}
				err = w.repository.UpsertGlobalTransaction(ctx, globalTx)
				if err != nil {
					w.logger.Error("cannot save redeemed tx", zap.Error(err))
				}
			case MethodUnkown:
				w.logger.Warn("method unkown", zap.String("tx", tx.Hash))
			}
			lastBlockNumberHex = tx.BlockNumber
		}

		newBlockNumber := int64(-1)
		if len(r.Result.Transactions) == 0 {
			newBlockNumber = lastBlock
		} else {
			lastBlockNumber := strings.Replace(lastBlockNumberHex, "0x", "", -1)
			newBlockNumber, err = strconv.ParseInt(lastBlockNumber, 16, 64)
			if err != nil {
				w.logger.Error("error parsing block number", zap.Error(err), zap.String("blockNumber", lastBlockNumber))
			}
		}

		w.logger.Debug("new block",
			zap.Int64("currentBlock", currentBlock),
			zap.Int64("lastBlock", lastBlock),
			zap.Int64("newBlockNumber", newBlockNumber),
			zap.String("lastBlockNumberHex", lastBlockNumberHex))

		if newBlockNumber != -1 {
			watcherBlock := storage.WatcherBlock{
				ID:          w.blockchain,
				BlockNumber: newBlockNumber,
				UpdatedAt:   time.Now(),
			}
			w.repository.UpdateWatcherBlock(ctx, watcherBlock)
		}

		pageToken = r.Result.NextPageToken
		if pageToken == "" {
			hasPage = false
		}
	}
}

func (w *EVMWatcher) Close() {
	close(w.close)
	w.wg.Wait()
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
	if len(input) < 10 {
		return MethodUnkown
	}
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
		return MethodUnkown

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

func remove0x(input string) string {
	return strings.Replace(input, "0x", "", -1)
}
