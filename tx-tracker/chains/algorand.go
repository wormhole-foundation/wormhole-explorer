package chains

import (
	"context"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

func fetchAlgorandTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxDetail, error) {

	// Decode txHash bytes
	h, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode from hex txHash=%s: %w", txHash, err)
	}

	// Encode as base32
	txid := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h)

	// Initialize an indexer client
	client, err := indexer.MakeClient(cfg.AlgorandBaseUrl, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create Algorand indexer client: %w", err)
	}

	// Fetch tx from the indexer API
	response, err := client.LookupTransaction(txid).Do(ctx)
	if err != nil {
		if strings.Contains(err.Error(), `404`) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to lookup Algorand txid=%s: %w", txid, err)
	}

	// Extract relevant fields and return
	txDetail := &TxDetail{
		Signer:       response.Transaction.Sender,
		Timestamp:    time.Unix(int64(response.Transaction.RoundTime), 0),
		NativeTxHash: response.Transaction.Id,
	}
	return txDetail, nil
}
