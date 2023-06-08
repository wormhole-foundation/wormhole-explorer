package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/terra"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Terra action methods.
const (
	MethodDepositTokens   = "deposit_tokens"
	MethodWithdrawTokens  = "withdraw_tokens"
	MethodRegisterAsset   = "register_asset"
	MethodContractUpgrade = "contract_upgrade"
	MethodCompleteWrapped = "complete_transfer_wrapped"
	MethodCompleteNative  = "complete_transfer_native"
	MethodCompleteTerra   = "complete_transfer_terra_native"
	MethodReplyHandler    = "reply_handler"
)

// Terrawatcher is a watcher for the terra chain.
type TerraWatcher struct {
	terraSDK        *terra.TerraSDK
	chainID         vaa.ChainID
	blockchain      string
	contractAddress string
	waitSeconds     uint16
	initialBlock    int64
	client          *http.Client
	repository      *storage.Repository
	logger          *zap.Logger
	close           chan bool
	wg              sync.WaitGroup
}

// TerraParams are the params for the terra watcher.
type TerraParams struct {
	ChainID         vaa.ChainID
	Blockchain      string
	ContractAddress string
	WaitSeconds     uint16
	InitialBlock    int64
}

// NewTerraWatcher creates a new terra watcher.
func NewTerraWatcher(terraSDK *terra.TerraSDK, params TerraParams, repository *storage.Repository, logger *zap.Logger) *TerraWatcher {
	return &TerraWatcher{
		terraSDK:        terraSDK,
		chainID:         params.ChainID,
		blockchain:      params.Blockchain,
		contractAddress: params.ContractAddress,
		waitSeconds:     params.WaitSeconds,
		initialBlock:    params.InitialBlock,
		client:          &http.Client{},
		repository:      repository,
		logger:          logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(params.ChainID))),
	}
}

// Start starts the terra watcher.
func (w *TerraWatcher) Start(ctx context.Context) error {
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
			w.logger.Info("clossing terra watcher by context")
			w.wg.Done()
			return nil
		case <-w.close:
			w.logger.Info("clossing terra watcher")
			w.wg.Done()
			return nil
		default:
			// get the latest block for the terra chain.
			lastBlock, err := w.terraSDK.GetLastBlock(ctx)
			if err != nil {
				w.logger.Error("cannot get terra lastblock", zap.Error(err))
			}

			// check if there are new blocks to process.
			if currentBlock < lastBlock {
				w.logger.Info("processing blocks", zap.Int64("from", currentBlock), zap.Int64("to", lastBlock))
				for block := currentBlock; block <= lastBlock; block++ {
					w.processBlock(ctx, block)
					// update block watcher
					watcherBlock := storage.WatcherBlock{
						ID:          w.blockchain,
						BlockNumber: block,
						UpdatedAt:   time.Now(),
					}
					w.repository.UpdateWatcherBlock(ctx, watcherBlock)
				}
			} else {
				w.logger.Info("waiting for new terra blocks")
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

func (w *TerraWatcher) Backfill(ctx context.Context, fromBlock uint64, toBlock uint64, pageSize uint64, persistBlock bool) {
	totalBlocks := getTotalBlocks(toBlock, fromBlock, pageSize)
	for i := uint64(0); i < totalBlocks; i++ {
		fromBlock, toBlock := getPage(fromBlock, i, pageSize, toBlock)
		w.logger.Info("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
		for block := fromBlock; block <= toBlock; block++ {
			w.processBlock(ctx, int64(block))
			if persistBlock {
				// update block watcher
				watcherBlock := storage.WatcherBlock{
					ID:          w.blockchain,
					BlockNumber: int64(block),
					UpdatedAt:   time.Now(),
				}
				w.repository.UpdateWatcherBlock(ctx, watcherBlock)
			}
		}
		w.logger.Info("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
	}
}

func (w *TerraWatcher) processBlock(ctx context.Context, block int64) {
	var offset *int
	hasPage := true
	for hasPage {

		// get transactions for the block.
		transactions, err := w.terraSDK.GetTransactionsByBlockHeight(ctx, block, offset)
		if err != nil {
			w.logger.Error("cannot get transactions by address", zap.Error(err), zap.Int64("block", block))
			time.Sleep(10 * time.Second)
			continue
		}

		// process all the transactions in the block
		for _, tx := range transactions.Txs {

			// unmarshall terra tx to wormhole token bridge tx.
			var wormholeTx terra.WormholeTerraTx
			txBytes, err := json.Marshal(tx.Tx)
			if err != nil {
				continue
			}
			err = json.Unmarshal(txBytes, &wormholeTx)
			if err != nil {
				continue
			}

			// check transaction contract address
			isTokenBridgeContract := w.checkTransactionContractAddress(wormholeTx)
			if !isTokenBridgeContract {
				continue
			}

			// unmarshall terra tx logs to wormhole token bridge logs.
			var wormholeTxLogs []terra.WormholeTerraTxLog
			txLogsBytes, err := json.Marshal(tx.Logs)
			if err != nil {
				w.logger.Debug("error marshall tx logs", zap.Error(err), zap.String("txHash", tx.Txhash),
					zap.Int64("block", block))
				continue
			}
			err = json.Unmarshal(txLogsBytes, &wormholeTxLogs)
			if err != nil {
				w.logger.Debug("error unmarshall to []terra.WormholeTerraLog", zap.Error(err),
					zap.String("txHash", tx.Txhash), zap.Int64("block", block))
				continue
			}

			// check transaction method
			supportedMethod, method := w.checkTransactionMethod(wormholeTxLogs)
			if !supportedMethod {
				continue
			}

			// get from, to and VAA from transaction message.
			from, to, vaa, err := w.getTransactionData(wormholeTx)
			if err != nil {
				w.logger.Error("cannot get transaction data", zap.Error(err),
					zap.String("txHash", tx.Txhash), zap.Int64("block", block))
				continue
			}

			if vaa == nil {
				w.logger.Error("cannot get VAA from transaction", zap.Error(err),
					zap.String("txHash", tx.Txhash), zap.Int64("block", block))
			}

			// create global transaction.
			updatedAt := time.Now()
			globalTx := storage.TransactionUpdate{
				ID: vaa.MessageID(),
				Destination: storage.DestinationTx{
					ChainID:     w.chainID,
					Status:      getStatus(tx),
					Method:      method,
					TxHash:      tx.Txhash,
					From:        from,
					To:          to,
					BlockNumber: strconv.Itoa(int(block)),
					Timestamp:   tx.Timestamp,
					UpdatedAt:   &updatedAt,
				},
			}

			// update global transaction and check if it should be updated.
			updateGlobalTransaction(ctx, globalTx, w.repository, w.logger)
		}

		if transactions.NextOffset == nil {
			hasPage = false
		} else {
			offset = transactions.NextOffset
		}
	}
}

func (w *TerraWatcher) checkTransactionContractAddress(tx terra.WormholeTerraTx) bool {
	for _, msg := range tx.Value.Msg {
		if msg.Value.Contract == w.contractAddress {
			return true
		}
	}
	return false
}

// checkTransactionMethod checks the method of the transaction.
// iterate over the logs, events and attributes to find the method.
func (w *TerraWatcher) checkTransactionMethod(tx []terra.WormholeTerraTxLog) (bool, string) {
	for _, log := range tx {
		for _, event := range log.Events {
			for _, attribute := range event.Attributes {
				if attribute.Key == "action" && filterTransactionMethod(attribute.Value) {
					return true, attribute.Value
				}
			}
		}
	}
	return false, ""
}

// getTransactionData
func (w *TerraWatcher) getTransactionData(tx terra.WormholeTerraTx) (string, string, *vaa.VAA, error) {
	for _, msg := range tx.Value.Msg {
		if msg.Value.Contract == w.contractAddress {
			// unmarshal vaa
			vaa, err := vaa.Unmarshal(msg.Value.ExecuteMsg.SubmitVaa.Data)
			if err != nil {
				return msg.Value.Sender, msg.Value.Contract, nil, err
			}
			return msg.Value.Sender, msg.Value.Contract, vaa, nil
		}
	}
	return "", "", nil, errors.New("cannot find transaction data")
}

func filterTransactionMethod(method string) bool {
	switch method {
	case MethodCompleteWrapped, MethodCompleteNative, MethodCompleteTerra:
		return true
	default:
		return false
	}
}

func getStatus(tx terra.Tx) string {
	if tx.Code == 0 {
		return domain.TxStatusConfirmed
	}
	return domain.TxStatusFailedToProcess
}

func (w *TerraWatcher) Close() {
	close(w.close)
	w.wg.Wait()
}
