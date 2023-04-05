package chains

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

type ethGetTransactionByHashResponse struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
}

type ethGetBlockByHashResponse struct {
	Timestamp string `json:"timestamp"`
	Number    string `json:"number"`
}

func fetchEthTx(
	ctx context.Context,
	txHash string,
	baseUrl string,
	apiKey string,
) (*TxDetail, error) {

	// build RPC URL
	url := baseUrl
	if apiKey != "" {
		url += "/" + apiKey
	}

	// initialize RPC client
	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// query transaction data
	var txReply ethGetTransactionByHashResponse
	err = client.CallContext(ctx, &txReply, "eth_getTransactionByHash", "0x"+txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx by hash: %w", err)
	}
	if txReply.BlockHash == "" || txReply.From == "" {
		return nil, errors.New("received empty response from the RPC service")
	}

	// query block data
	blkParams := []interface{}{
		txReply.BlockHash, // tx hash
		false,             // include transactions?
	}
	var blkReply ethGetBlockByHashResponse
	err = client.CallContext(ctx, &blkReply, "eth_getBlockByHash", blkParams...)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	// parse transaction timestamp
	timestamp, err := timestampFromHex(blkReply.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block timestamp: %w", err)
	}

	// build results and return
	txDetail := &TxDetail{
		Signer:       strings.ToLower(txReply.From),
		Timestamp:    timestamp,
		NativeTxHash: fmt.Sprintf("0x%s", strings.ToLower(txHash)),
	}
	return txDetail, nil
}
