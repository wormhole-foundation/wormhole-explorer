package watcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/support"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	//Method names
	MethodCompleteTransfer     = "completeTransfer"
	MethodWrapAndTransfer      = "wrapAndTransfer"
	MethodTransferTokens       = "transferTokens"
	MethodAttestToken          = "attestToken"
	MethodCompleteAndUnwrapETH = "completeAndUnwrapETH"
	MethodCreateWrapped        = "createWrapped"
	MethodUpdateWrapped        = "updateWrapped"
	MethodUnkown               = "unknown"

	//Method ids
	MethodIDCompleteTransfer     = "0xc6878519"
	MethodIDWrapAndTransfer      = "0x9981509f"
	MethodIDTransferTokens       = "0x0f5287b0"
	MethodIDAttestToken          = "0xc48fa115"
	MethodIDCompleteAndUnwrapETH = "0xff200cde"
	MethodIDCreateWrapped        = "0xe8059810"
	MethodIDUpdateWrapped        = "0xf768441f"

	//Transaction status
	TxStatusSuccess      = "0x1"
	TxStatusFailReverted = "0x0"
)

type EVMParams struct {
	ChainID         vaa.ChainID
	Blockchain      string
	ContractAddress string
	SizeBlocks      uint8
	WaitSeconds     uint16
	InitialBlock    int64
}

type EvmTransaction struct {
	Hash           string
	From           string
	To             string
	Status         string
	BlockNumber    string
	BlockTimestamp string
	Input          string
}

// get transaction status
func getTxStatus(status string) string {
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
func getMethodByInput(input string) string {
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

func processTransaction(ctx context.Context, chainID vaa.ChainID, tx *EvmTransaction, repository *storage.Repository, logger *zap.Logger) {
	method := getMethodByInput(tx.Input)

	log := logger.With(
		zap.String("txHash", tx.Hash),
		zap.String("method", method),
		zap.String("block", tx.BlockNumber))
	log.Debug("new tx")

	switch method {
	case MethodCompleteTransfer, MethodCompleteAndUnwrapETH, MethodCreateWrapped, MethodUpdateWrapped:

		vaa, err := parseInput(tx.Input)
		if err != nil {
			log.Error("cannot parse VAA", zap.Error(err))
			return
		}

		updatedAt := time.Now()
		globalTx := storage.TransactionUpdate{
			ID: vaa.MessageID(),
			Destination: storage.DestinationTx{
				ChainID:     chainID,
				Status:      getTxStatus(tx.Status),
				Method:      getMethodByInput(tx.Input),
				TxHash:      support.Remove0x(tx.Hash),
				To:          tx.To,
				From:        tx.From,
				BlockNumber: getBlockNumber(tx.BlockNumber, log),
				Timestamp:   getTimestamp(tx.BlockTimestamp, log),
				UpdatedAt:   &updatedAt,
			},
		}
		err = repository.UpsertGlobalTransaction(ctx, globalTx)
		if err != nil {
			log.Error("cannot save redeemed tx", zap.Error(err))
		} else {
			log.Info("saved redeemed tx", zap.String("vaa", vaa.MessageID()))

		}
	case MethodUnkown:
		log.Debug("method unkown")
	}
}
