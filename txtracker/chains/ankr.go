package chains

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

type ankrGetTransactionByHashParams struct {
	TransactionHash string `json:"transactionHash"`
	DecodeLogs      bool   `json:"decodeLogs"`
	DecodeTxData    bool   `json:"decodeTxData"`
}

type ankrGetTransactionsByHashResponse struct {
	Transactions []ankrTransaction `json:"transactions"`
}

type ankrTransaction struct {
	From      string    `json:"from"`
	Timestamp string    `json:"timestamp"`
	Logs      []ankrLog `json:"logs"`
}

type ankrLog struct {
	Event ankrEvent `json:"event"`
}

type ankrEvent struct {
	Name   string           `json:"name"`
	Inputs []ankrEventInput `json:"inputs"`
}

type ankrEventInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         uint64 `json:"size"`
	ValueDecoded string `json:"valueDecoded"`
}

func ankrFetchBscTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxDetail, error) {
	return ankrFetchTx(ctx, cfg, TokenBridgeBsc, txHash)
}

func ankrFetchEthTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxDetail, error) {
	return ankrFetchTx(ctx, cfg, TokenBridgeEthereum, txHash)
}

func ankrFetchPolygonTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxDetail, error) {
	return ankrFetchTx(ctx, cfg, TokenBridgePolygon, txHash)
}

func ankrFetchTx(
	ctx context.Context,
	cfg *config.Settings,
	tokenBridgeAddr string,
	txHash string,
) (*TxDetail, error) {

	// initialize RPC client
	client, err := rpc.DialContext(ctx, cfg.AnkrBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// query transaction data
	args := &ankrGetTransactionByHashParams{
		TransactionHash: "0x" + txHash,
		DecodeLogs:      true,
		DecodeTxData:    true,
	}
	var reply ankrGetTransactionsByHashResponse
	err = client.CallContext(ctx, &reply, "ankr_getTransactionsByHash", args)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx by hash: %w", err)
	}

	// iterate transaction logs
	var txDetail *TxDetail
	for i := range reply.Transactions {
		for j := range reply.Transactions[i].Logs {

			ev := &reply.Transactions[i].Logs[j].Event

			if ev.Name == "Transfer" && len(ev.Inputs) == 3 {

				// validate sender
				if ev.Inputs[0].Name != "from" {
					return nil, fmt.Errorf(`expected input name to be "from", but encountered: %s`, ev.Inputs[0].Name)
				}
				source := ev.Inputs[0].ValueDecoded

				// validate receiver
				if ev.Inputs[1].Name != "to" {
					return nil, fmt.Errorf(`expected input name to be "to", but encountered: %s`, ev.Inputs[1].Name)
				}
				destination := strings.ToLower(ev.Inputs[1].ValueDecoded)

				// validate timestamp
				hexDigits := strings.Replace(reply.Transactions[i].Timestamp, "0x", "", 1)
				hexDigits = strings.Replace(hexDigits, "0X", "", 1)
				epoch, err := strconv.ParseInt(hexDigits, 16, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse transaction timestamp")
				}
				timestamp := time.Unix(epoch, 0)

				// make sure the transfer is interacting with the token bridge
				if destination != tokenBridgeAddr {
					continue
				}

				// set the result
				if txDetail != nil {
					return nil, fmt.Errorf("encountered more than one transfer/deposit event")
				}
				txDetail = &TxDetail{
					Timestamp:   timestamp,
					Source:      source,
					Destination: destination,
				}

			} else if ev.Name == "Deposit" && len(ev.Inputs) == 2 {

				// set sender
				source := strings.ToLower(reply.Transactions[i].From)

				// validate receiver
				if ev.Inputs[0].Name != "account" {
					return nil, fmt.Errorf(`expected input name to be "account", but encountered: %s`, ev.Inputs[0].Name)
				}
				destination := strings.ToLower(ev.Inputs[0].ValueDecoded)

				// validate amount
				if ev.Inputs[1].Name != "amount" {
					return nil, fmt.Errorf(`expected input name to be "amount", but encountered: %s`, ev.Inputs[1].Name)
				}
				amount := big.NewInt(0)
				_, ok := amount.SetString(ev.Inputs[1].ValueDecoded, 10)
				if !ok {
					return nil, fmt.Errorf("failed to parse amount")
				}

				// validate timestamp
				hexDigits := strings.Replace(reply.Transactions[i].Timestamp, "0x", "", 1)
				hexDigits = strings.Replace(hexDigits, "0X", "", 1)
				epoch, err := strconv.ParseInt(hexDigits, 16, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse transaction timestamp")
				}
				timestamp := time.Unix(epoch, 0)

				// make sure the transfer is interacting with the token bridge
				if destination != tokenBridgeAddr {
					continue
				}

				// set the result
				if txDetail != nil {
					return nil, fmt.Errorf("encountered more than one transfer/deposit event")
				}
				txDetail = &TxDetail{
					Timestamp:   timestamp,
					Source:      source,
					Destination: destination,
				}
			}

		}
	}
	if txDetail == nil {
		return nil, fmt.Errorf("expected at least one transfer/deposit event")
	}

	return txDetail, nil
}
