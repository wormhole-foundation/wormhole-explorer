package watcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/aptos"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const CompleteTransferMethod = "complete_transfer::submit_vaa_and_register_entry"

const aptosMaxRetries = 10
const aptosRetryDelay = 5 * time.Second

type AptosParams struct {
	Blockchain      string
	ContractAddress string
	SizeBlocks      uint8
	WaitSeconds     uint16
	InitialBlock    int64
}

type AptosWatcher struct {
	client          *aptos.AptosSDK
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
	metrics         metrics.Metrics
}

func NewAptosWatcher(client *aptos.AptosSDK, params AptosParams, repo *storage.Repository, metrics metrics.Metrics, logger *zap.Logger) *AptosWatcher {
	chainID := vaa.ChainIDAptos
	return &AptosWatcher{
		client:          client,
		chainID:         chainID,
		blockchain:      params.Blockchain,
		contractAddress: params.ContractAddress,
		sizeBlocks:      params.SizeBlocks,
		waitSeconds:     params.WaitSeconds,
		initialBlock:    params.InitialBlock,
		repository:      repo,
		metrics:         metrics,
		logger:          logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(chainID))),
	}
}

func (w *AptosWatcher) Start(ctx context.Context) error {
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
			maxBlocks := uint64(w.sizeBlocks)
			w.logger.Debug("current block", zap.Uint64("current", currentBlock), zap.Uint64("last", lastBlock))
			w.metrics.SetLastBlock(w.chainID, lastBlock)
			if currentBlock < lastBlock {
				totalBlocks := (lastBlock-currentBlock)/maxBlocks + 1
				for i := 0; i < int(totalBlocks); i++ {
					fromBlock := currentBlock + uint64(i)*maxBlocks
					toBlock := fromBlock + maxBlocks - 1
					if toBlock > lastBlock {
						toBlock = lastBlock
					}
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

func (w *AptosWatcher) Close() {
	close(w.close)
	w.wg.Wait()
}

func (w *AptosWatcher) Backfill(ctx context.Context, fromBlock uint64, toBlock uint64, pageSize uint64, persistBlock bool) {
	totalBlocks := getTotalBlocks(toBlock, fromBlock, pageSize)
	for i := uint64(0); i < totalBlocks; i++ {
		fromBlock, toBlock := getPage(fromBlock, i, pageSize, toBlock)
		w.logger.Info("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
		w.processBlock(ctx, fromBlock, toBlock, persistBlock)
		w.logger.Info("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
	}
}

func (w *AptosWatcher) processBlock(ctx context.Context, fromBlock uint64, toBlock uint64, updateWatcherBlock bool) {

	for block := fromBlock; block <= toBlock; block++ {
		w.logger.Debug("processing block", zap.Uint64("block", block))
		retry.Do(
			func() error {
				// get the transactions for the block.
				result, err := w.client.GetBlock(ctx, block)
				if err != nil {
					w.logger.Error("cannot get block", zap.Uint64("block", block), zap.Error(err))
					if err == aptos.ErrTooManyRequests {
						return err
					}
					return nil
				}
				blockTime, err := result.GetBlockTime()
				if err != nil {
					w.logger.Warn("cannot get block time", zap.Uint64("block", block), zap.Error(err))
				}

				for _, tx := range result.Transactions {
					w.processTransaction(ctx, tx, block, blockTime)
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
			retry.Attempts(aptosMaxRetries),
			retry.Delay(aptosRetryDelay),
		)
	}
}

func (w *AptosWatcher) processTransaction(ctx context.Context, tx aptos.Transaction, block uint64, blockTime *time.Time) {

	found, method := w.isTokenBridgeFunction(tx.Payload.Function)
	if !found {
		return
	}

	log := w.logger.With(
		zap.String("txHash", tx.Hash),
		zap.String("txVersion", tx.Version),
		zap.String("function", tx.Payload.Function),
		zap.Uint64("block", block))

	if method != CompleteTransferMethod {
		log.Warn("unkown method", zap.String("method", method))
		return
	}

	log.Debug("found Wormhole transaction")

	if len(tx.Payload.Arguments) != 1 {
		log.Error("invalid number of arguments",
			zap.Int("arguments", len(tx.Payload.Arguments)))
		return
	}

	switch tx.Payload.Arguments[0].(type) {
	case string:
	default:
		log.Error("invalid type of argument")
		return
	}

	vaaArg := tx.Payload.Arguments[0].(string)
	data, err := hex.DecodeString(strings.TrimPrefix(vaaArg, "0x"))
	if err != nil {
		log.Error("invalid vaa argument",
			zap.String("argument", vaaArg),
			zap.Error(err))
		return
	}

	result, err := vaa.Unmarshal(data)
	if err != nil {
		log.Error("invalid vaa",
			zap.Error(err))
		return
	}

	txResult, err := w.client.GetTransaction(ctx, tx.Version)
	if err != nil {
		log.Error("get transaction error",
			zap.String("version", tx.Version),
			zap.Error(err))
		return
	}
	status := domain.DstTxStatusFailedToProcess
	if txResult.Success {
		status = domain.DstTxStatusConfirmed
	}
	updatedAt := time.Now()
	globalTx := storage.TransactionUpdate{
		ID: result.MessageID(),
		Destination: storage.DestinationTx{
			ChainID:     w.chainID,
			Status:      status,
			Method:      method,
			TxHash:      tx.Hash,
			BlockNumber: strconv.FormatUint(block, 10),
			Timestamp:   blockTime,
			UpdatedAt:   &updatedAt,
		},
	}

	// update global transaction and check if it should be updated.
	updateGlobalTransaction(ctx, w.chainID, globalTx, w.repository, log)
}

func (w *AptosWatcher) isTokenBridgeFunction(fn string) (bool, string) {
	prefixFunction := fmt.Sprintf("%s::", w.contractAddress)
	if !strings.HasPrefix(fn, prefixFunction) {
		return false, ""
	}

	return true, strings.TrimPrefix(fn, prefixFunction)
}
