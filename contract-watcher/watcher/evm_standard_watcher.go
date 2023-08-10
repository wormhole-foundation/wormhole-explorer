package watcher

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/evm"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	evmMaxRetries = 10
	evmRetryDelay = 5 * time.Second
)

type EvmStandardWatcher struct {
	client           *evm.EvmSDK
	chainID          vaa.ChainID
	blockchain       string
	contractAddress  []string
	methodsByAddress map[string][]config.BlockchainMethod
	maxBlocks        uint64
	waitSeconds      uint16
	initialBlock     int64
	repository       *storage.Repository
	logger           *zap.Logger
	close            chan bool
	wg               sync.WaitGroup
	metrics          metrics.Metrics
}

func NewEvmStandardWatcher(client *evm.EvmSDK, params EVMParams, repo *storage.Repository, metrics metrics.Metrics, logger *zap.Logger) *EvmStandardWatcher {
	addresses := make([]string, 0, len(params.MethodsByAddress))
	for address := range params.MethodsByAddress {
		addresses = append(addresses, address)
	}
	return &EvmStandardWatcher{
		client:           client,
		chainID:          params.ChainID,
		blockchain:       params.Blockchain,
		contractAddress:  addresses,
		methodsByAddress: params.MethodsByAddress,
		maxBlocks:        uint64(params.SizeBlocks),
		waitSeconds:      params.WaitSeconds,
		initialBlock:     params.InitialBlock,
		repository:       repo,
		metrics:          metrics,
		logger:           logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(params.ChainID))),
	}
}

func (w *EvmStandardWatcher) Start(ctx context.Context) error {
	// get the current block for the chain.
	cBlock, err := w.repository.GetCurrentBlock(ctx, w.blockchain, w.initialBlock)
	if err != nil {
		w.logger.Error("cannot get current block", zap.Error(err))
		return err
	}
	currentBlock := uint64(cBlock)
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
			lastBlock, err := w.client.GetLatestBlock(ctx)
			if err != nil {
				w.logger.Error("cannot get latest block", zap.Error(err))
			}
			w.logger.Debug("current block", zap.Uint64("current", currentBlock), zap.Uint64("last", lastBlock))

			if currentBlock < lastBlock {
				w.metrics.SetLastBlock(w.chainID, lastBlock)
				totalBlocks := getTotalBlocks(lastBlock, currentBlock, w.maxBlocks)
				for i := uint64(0); i < totalBlocks; i++ {
					fromBlock, toBlock := getPage(currentBlock, i, w.maxBlocks, lastBlock)
					w.logger.Debug("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
					w.processBlock(ctx, fromBlock, toBlock, true)
					w.logger.Debug("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
				}
				// process all the blocks between current and last block.
			} else {
				w.logger.Debug("waiting for new blocks")
				select {
				case <-ctx.Done():
					w.wg.Done()
					return nil
				case <-time.After(time.Duration(w.waitSeconds) * time.Second):
				}
			}
			if lastBlock > currentBlock {
				currentBlock = lastBlock
			}
		}
	}

}

func (w *EvmStandardWatcher) Backfill(ctx context.Context, fromBlock uint64, toBlock uint64, pageSize uint64, persistBlock bool) {
	totalBlocks := getTotalBlocks(toBlock, fromBlock, pageSize)
	for i := uint64(0); i < totalBlocks; i++ {
		fromBlock, toBlock := getPage(fromBlock, i, pageSize, toBlock)
		w.logger.Info("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
		w.processBlock(ctx, fromBlock, toBlock, persistBlock)
		w.logger.Info("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
	}
}

func (w *EvmStandardWatcher) processBlock(ctx context.Context, fromBlock uint64, toBlock uint64, updateWatcherBlock bool) {
	for block := fromBlock; block <= toBlock; block++ {
		w.logger.Debug("processing block", zap.Uint64("block", block))
		retry.Do(
			func() error {
				// get the transactions for the block.
				blockResult, err := w.client.GetBlock(ctx, block)
				if err != nil {
					w.logger.Error("cannot get block", zap.Uint64("block", block), zap.Error(err))
					if err == evm.ErrTooManyRequests {
						return err
					}
					return nil
				}

				for _, tx := range blockResult.Transactions {

					// only process transactions to the contract address.
					_, ok := w.methodsByAddress[strings.ToLower(tx.To)]
					if !ok {
						continue
					}

					evmTx := &EvmTransaction{
						Hash: tx.Hash,
						From: tx.From,
						To:   tx.To,
						Status: func() (string, error) {
							var status string
							// add retry to get the transaction receipt.
							err := retry.Do(
								func() error {
									tranctionReceipt, err := w.client.GetTransactionReceipt(ctx, tx.Hash)
									if err != nil {
										w.logger.Error("cannot get tranction receipt",
											zap.Uint64("block", block),
											zap.String("txHash", tx.Hash),
											zap.Error(err))
										if err == evm.ErrTooManyRequests {
											return err
										}
										return nil
									}
									// get the status of the transaction
									status = tranctionReceipt.Status
									return nil
								},
								retry.Attempts(evmMaxRetries),
								retry.Delay(evmRetryDelay),
							)
							return status, err
						},
						BlockNumber:    tx.BlockNumber,
						BlockTimestamp: blockResult.Timestamp,
						Input:          tx.Input,
					}
					processTransaction(ctx, w.chainID, evmTx, w.methodsByAddress, w.repository, w.logger)
				}

				if updateWatcherBlock {
					// update the last block number processed in the database.
					watcherBlock := storage.WatcherBlock{
						ID:          w.blockchain,
						BlockNumber: int64(block),
						UpdatedAt:   time.Now(),
					}
					return w.repository.UpdateWatcherBlock(ctx, w.chainID, watcherBlock)
				}
				return nil
			},
			retry.Attempts(evmMaxRetries),
			retry.Delay(evmRetryDelay),
		)
	}
}

func (w *EvmStandardWatcher) Close() {
	close(w.close)
	w.wg.Wait()
}
