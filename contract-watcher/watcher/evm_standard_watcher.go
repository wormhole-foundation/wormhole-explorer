package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/evm"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	evmMaxRetries = 10
	evmRetryDelay = 5 * time.Second
)

type EvmStandarWatcher struct {
	client          *evm.EvmSDK
	chainID         vaa.ChainID
	blockchain      string
	contractAddress string
	maxBlocks       uint64
	waitSeconds     uint16
	initialBlock    int64
	repository      *storage.Repository
	logger          *zap.Logger
	close           chan bool
	wg              sync.WaitGroup
}
type EvmStandarParams struct {
	ChainID         vaa.ChainID
	Blockchain      string
	ContractAddress string
	SizeBlocks      uint8
	WaitSeconds     uint16
	InitialBlock    uint64
}

func NewEvmStandarWatcher(client *evm.EvmSDK, params EVMParams, repo *storage.Repository, logger *zap.Logger) *EvmStandarWatcher {
	return &EvmStandarWatcher{
		client:          client,
		chainID:         params.ChainID,
		blockchain:      params.Blockchain,
		contractAddress: params.ContractAddress,
		maxBlocks:       uint64(params.SizeBlocks),
		waitSeconds:     params.WaitSeconds,
		initialBlock:    params.InitialBlock,
		repository:      repo,
		logger:          logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(params.ChainID))),
	}
}

func (w *EvmStandarWatcher) Start(ctx context.Context) error {
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
			w.logger.Info("current block", zap.Uint64("current", currentBlock), zap.Uint64("last", lastBlock))
			if currentBlock < lastBlock {
				totalBlocks := getTotalBlocks(lastBlock, currentBlock, w.maxBlocks)
				for i := uint64(0); i < totalBlocks; i++ {
					fromBlock, toBlock := getPage(currentBlock, i, w.maxBlocks, lastBlock)
					w.logger.Info("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
					w.processBlock(ctx, fromBlock, toBlock)
					w.logger.Info("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
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

func (w *EvmStandarWatcher) processBlock(ctx context.Context, fromBlock uint64, toBlock uint64) {
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
					if w.contractAddress != tx.To {
						continue
					}

					evmTx := &EvmTransaction{
						Hash:           tx.Hash,
						From:           tx.From,
						To:             tx.To,
						Status:         TxStatusSuccess,
						BlockNumber:    tx.BlockNumber,
						BlockTimestamp: blockResult.Timestamp,
						Input:          tx.Input,
					}
					processTransaction(ctx, w.chainID, evmTx, w.repository, w.logger)
				}
				// update the last block number processed in the database.
				watcherBlock := storage.WatcherBlock{
					ID:          w.blockchain,
					BlockNumber: int64(block),
					UpdatedAt:   time.Now(),
				}
				return w.repository.UpdateWatcherBlock(ctx, watcherBlock)
			},
			retry.Attempts(evmMaxRetries),
			retry.Delay(evmRetryDelay),
		)
	}
}

func (w *EvmStandarWatcher) Close() {
	close(w.close)
	w.wg.Wait()
}
