package connectors

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

func FetchBscTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxData, error) {

	// initialize RPC client
	client, err := rpc.DialContext(ctx, cfg.AnkrBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// query transaction data
	args := &ankrGetTransactionByHashParams{
		TransactionHash: txHash,
		DecodeLogs:      true,
		DecodeTxData:    true,
	}
	var reply ankrGetTransactionsByHashResponse
	err = client.CallContext(ctx, &reply, "ankr_getTransactionsByHash", args)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx by hash: %w", err)
	}

	// iterate transaction logs
	var found bool
	var txData TxData
	for i := range reply.Transactions {
		for j := range reply.Transactions[i].Logs {

			ev := &reply.Transactions[i].Logs[j].Event

			if ev.Name == "Transfer" && len(ev.Inputs) == 3 {

				if found {
					return nil, fmt.Errorf("encountered more than one transfer/deposit event")
				}
				found = true

				// validate sender
				if ev.Inputs[0].Name != "from" {
					return nil, fmt.Errorf(`expected input name to be "from", but encountered: %s`, ev.Inputs[0].Name)
				}
				txData.Source = ev.Inputs[0].ValueDecoded

				// validate receiver
				if ev.Inputs[1].Name != "to" {
					return nil, fmt.Errorf(`expected input name to be "to", but encountered: %s`, ev.Inputs[1].Name)
				}
				txData.Destination = strings.ToLower(ev.Inputs[1].ValueDecoded)

				// validate amount
				if ev.Inputs[2].Name != "value" {
					return nil, fmt.Errorf(`expected input name to be "value", but encountered: %s`, ev.Inputs[2].Name)
				}
				txData.Amount = big.NewInt(0)
				_, ok := txData.Amount.SetString(ev.Inputs[2].ValueDecoded, 10)
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
				txData.Date = time.Unix(epoch, 0)

			} else if ev.Name == "Deposit" && len(ev.Inputs) == 2 {

				if found {
					return nil, fmt.Errorf("encountered more than one transfer/deposit event")
				}
				found = true

				// set sender
				txData.Source = strings.ToLower(reply.Transactions[i].From)

				// validate receiver
				if ev.Inputs[0].Name != "account" {
					return nil, fmt.Errorf(`expected input name to be "account", but encountered: %s`, ev.Inputs[0].Name)
				}
				txData.Destination = strings.ToLower(ev.Inputs[0].ValueDecoded)

				// validate amount
				if ev.Inputs[1].Name != "amount" {
					return nil, fmt.Errorf(`expected input name to be "amount", but encountered: %s`, ev.Inputs[1].Name)
				}
				txData.Amount = big.NewInt(0)
				_, ok := txData.Amount.SetString(ev.Inputs[1].ValueDecoded, 10)
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
				txData.Date = time.Unix(epoch, 0)
			}

		}
	}
	if !found {
		return nil, fmt.Errorf("expected at least one transfer/deposit event")
	}

	return &txData, nil
}
