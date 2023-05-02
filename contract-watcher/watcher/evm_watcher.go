package watcher

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
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
	maxBlocks := int64(w.sizeBlocks)
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
			lastBlock := currentBlock
			if len(stats.Result.Stats) > 0 {
				lastBlock = stats.Result.Stats[0].LatestBlockNumber
			}

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
			evmTx := &EvmTransaction{
				Hash: tx.Hash,
				From: tx.From,
				To:   tx.To,
				Status: func() (string, error) {
					return tx.Status, nil
				},
				BlockNumber:    tx.BlockNumber,
				BlockTimestamp: tx.Timestamp,
				Input:          tx.Input,
			}
			processTransaction(ctx, w.chainID, evmTx, w.repository, w.logger)
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
