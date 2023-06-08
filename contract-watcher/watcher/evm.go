package watcher

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/support"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	//Transaction status
	TxStatusSuccess      = "0x1"
	TxStatusFailReverted = "0x0"
)

type EVMParams struct {
	ChainID          vaa.ChainID
	Blockchain       string
	SizeBlocks       uint8
	WaitSeconds      uint16
	InitialBlock     int64
	MethodsByAddress map[string][]config.BlockchainMethod
}

type EVMAddressesParams struct {
	ChainID      vaa.ChainID
	Blockchain   string
	SizeBlocks   uint8
	WaitSeconds  uint16
	InitialBlock int64
}

type EvmGetStatusFunc func() (string, error)

type EvmTransaction struct {
	Hash           string
	From           string
	To             string
	Status         EvmGetStatusFunc
	BlockNumber    string
	BlockTimestamp string
	Input          string
}

// get gloabal transaction status from evm blockchain status code.
func getTxStatus(status string) string {
	switch status {
	case TxStatusSuccess:
		return domain.DstTxStatusConfirmed
	case TxStatusFailReverted:
		return domain.DstTxStatusFailedToProcess
	default:
		return domain.DstTxStatusUnkonwn
	}
}

// get method ID from transaction input
func getMethodIDByInput(input string) string {
	if len(input) < 10 {
		return config.MethodUnkown
	}
	return input[0:10]
}

// get the input and extract the method signature and VAA
func parseInput(input string) (*vaa.VAA, error) {
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

func getBlockNumber(s string, logger *zap.Logger) string {
	value, err := strconv.ParseInt(support.Remove0x(s), 16, 64)
	if err != nil {
		logger.Error("cannot convert to int", zap.Error(err))
		return s
	}
	return strconv.FormatInt(value, 10)
}

func getTimestamp(s string, logger *zap.Logger) *time.Time {
	value, err := strconv.ParseInt(support.Remove0x(s), 16, 64)
	if err != nil {
		logger.Error("cannot convert to timestamp", zap.Error(err))
		return nil
	}
	tm := time.Unix(value, 0)
	return &tm
}

func processTransaction(ctx context.Context, chainID vaa.ChainID, tx *EvmTransaction, methodsByAddress map[string][]config.BlockchainMethod, repository *storage.Repository, logger *zap.Logger) {
	// get methodID from the transaction.
	txMethod := getMethodIDByInput(tx.Input)

	log := logger.With(
		zap.String("txHash", tx.Hash),
		zap.String("method", txMethod),
		zap.String("block", tx.BlockNumber))
	log.Debug("new tx")

	// get methods by address.
	methods, ok := methodsByAddress[strings.ToLower(tx.To)]
	if !ok {
		log.Debug("method unkown")
		return
	}

	for _, method := range methods {
		if method.ID == txMethod {
			// get vaa from transaction input
			vaa, err := parseInput(tx.Input)
			if err != nil {
				log.Error("cannot parse VAA", zap.Error(err))
				return

			}

			// get evm blockchain status code
			txStatusCode, err := tx.Status()
			if err != nil {
				log.Error("cannot get tx status", zap.Error(err))
				return
			}

			updatedAt := time.Now()
			globalTx := storage.TransactionUpdate{
				ID: vaa.MessageID(),
				Destination: storage.DestinationTx{
					ChainID:     chainID,
					Status:      getTxStatus(txStatusCode),
					Method:      method.Name,
					TxHash:      support.Remove0x(tx.Hash),
					To:          tx.To,
					From:        tx.From,
					BlockNumber: getBlockNumber(tx.BlockNumber, log),
					Timestamp:   getTimestamp(tx.BlockTimestamp, log),
					UpdatedAt:   &updatedAt,
				},
			}

			// update global transaction and check if it should be updated.
			updateGlobalTransaction(ctx, globalTx, repository, log)
			break
		}
	}
}
