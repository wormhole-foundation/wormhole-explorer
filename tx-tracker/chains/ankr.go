package chains

import (
	"context"
	"fmt"
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
	To        string    `json:"to"`
	Timestamp string    `json:"timestamp"`
	Logs      []ankrLog `json:"logs"`
	Status    string    `json:"status"`
}

type ankrLog struct {
	Event  ankrEvent `json:"event"`
	Topics []string  `json:"topics"`
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

func ankrFetchTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	txHash string,
) (*TxDetail, error) {

	// build RPC URL
	url := cfg.AnkrBaseUrl
	if cfg.AnkrApiKey != "" {
		url += "/" + cfg.AnkrApiKey
	}

	// initialize RPC client
	client, err := rpc.DialContext(ctx, url)
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

	// make sure we got exactly one transaction
	if len(reply.Transactions) == 0 {
		return nil, ErrTransactionNotFound
	} else if len(reply.Transactions) > 1 {
		return nil, fmt.Errorf("expected one transaction for txid=%s, but found %d", txHash, len(reply.Transactions))
	}

	// parse transaction timestamp
	var timestamp time.Time
	{
		hexDigits := strings.Replace(reply.Transactions[0].Timestamp, "0x", "", 1)
		hexDigits = strings.Replace(hexDigits, "0X", "", 1)
		epoch, err := strconv.ParseInt(hexDigits, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction timestamp: %w", err)
		}
		timestamp = time.Unix(epoch, 0).UTC()
	}

	// build results and return
	txDetail := &TxDetail{
		Signer:       strings.ToLower(reply.Transactions[0].From),
		Timestamp:    timestamp,
		NativeTxHash: fmt.Sprintf("0x%s", strings.ToLower(txHash)),
	}
	return txDetail, nil
}
