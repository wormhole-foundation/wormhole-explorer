package watchers

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/event-watcher/clients"
	"go.uber.org/zap"
	"golang.org/x/exp/constraints"
)

const bulkSize = 100

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

type EvmWatcher struct {
	logger              *zap.Logger
	db                  *dbutil.Session
	client              *clients.EthRpcClient
	coreContractAddress string
	logTopic            string
}

func NewEvmWatcher(
	logger *zap.Logger,
	db *dbutil.Session,
	coreContractAddress string,
	logTopic string,
	url string,
	auth string,
) *EvmWatcher {

	w := EvmWatcher{
		logger:              logger,
		db:                  db,
		client:              clients.NewEthRpcClient(url, auth),
		coreContractAddress: coreContractAddress,
		logTopic:            logTopic,
	}

	return &w
}

func (w *EvmWatcher) Watch(ctx context.Context) {

	//TODO:
	// - initialize current block in the database, if not already initialized.
	// - get current block from database
	var currentBlock uint64 = 0

	for {
		// Get the current blockchain head
		latestBlock, err := w.client.GetBlockNumber(ctx)
		if err != nil {
			w.logger.Error("failed to get latest block number",
				zap.String("url", w.client.Url),
				zap.Error(err),
			)
			continue
		}

		// Process blocks in bulk
		for currentBlock < latestBlock {
			from := currentBlock
			to := min(currentBlock+bulkSize, latestBlock)
			w.processBlockRange(ctx, from, to)

			currentBlock = latestBlock
		}
	}
}

func (w *EvmWatcher) processBlockRange(ctx context.Context, fromBlock uint64, toBlock uint64) {

	var logs []clients.Log
	var err error

	// Retry until success
	for {
		logs, err = w.client.GetLogs(ctx, fromBlock, toBlock, w.coreContractAddress, w.logTopic)
		if err != nil {
			w.logger.Error("failed to get logs",
				zap.String("url", w.client.Url),
				zap.String("coreContractAddress", w.coreContractAddress),
				zap.String("topic", w.logTopic),
				zap.Uint64("fromBlock", fromBlock),
				zap.Uint64("toBlock", toBlock),
				zap.Error(err),
			)
		}
		break
	}

	// Process logs
	// TODO:
	// - update current block in database
	// - fire events for other services
	for i := range logs {
		log := logs[i]
		w.logger.Info("found log", zap.String("transactionHash", log.TransactionHash))
	}
}
