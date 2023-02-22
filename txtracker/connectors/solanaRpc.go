package connectors

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/mr-tron/base58"
	"github.com/portto/solana-go-sdk/client"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

func FetchSolanaTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxData, error) {

	c := client.NewClient(cfg.SolanaBaseUrl)

	// Decode txHash bytes
	h, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode from hex txHash=%s: %w", txHash, err)
	}

	// Get transaction signatures for the given account
	sigs, err := c.GetSignaturesForAddress(ctx, base58.Encode(h))
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures for txHash=%s: %w", txHash, err)
	}
	if len(sigs) != 1 {
		return nil, fmt.Errorf("expected account to have exactly one signature, but found %d", len(sigs))
	}

	// Fetch the portal token bridge transaction
	tx, err := c.GetTransaction(ctx, sigs[0].Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction sig=%s, err: %v", sigs[0].Signature, err)
	}

	// Check preconditions
	//
	// The transaction data structure uses lots of parallel arrays.
	// Here we're making sure that we won't get an error due to an index being out of bounds.
	if len(tx.Meta.PostBalances) != len(tx.Meta.PreBalances) {
		return nil, fmt.Errorf("encountered mismatching sizes in pre/post balance arrays")
	}
	for i := range tx.Meta.PreTokenBalances {
		if tx.Meta.PreTokenBalances[i].AccountIndex != tx.Meta.PostTokenBalances[i].AccountIndex {
			return nil, fmt.Errorf("mismatching account indexes")
		}
		if tx.Meta.PreTokenBalances[i].AccountIndex >= uint64(len(tx.Transaction.Message.Accounts)) {
			return nil, fmt.Errorf("pre-token balance index out of range")
		}
		if tx.Meta.PostTokenBalances[i].AccountIndex >= uint64(len(tx.Transaction.Message.Accounts)) {
			return nil, fmt.Errorf("post-token balance index out of range")
		}
	}

	// Initialize the struct containing resuts
	var txData TxData
	if tx.BlockTime != nil {
		txData.Timestamp = time.Unix(*tx.BlockTime, 0)
	}

	// Iterate through balances changes to find the funds source and destination.
	var receiverFound, senderFound bool
	for i := range tx.Meta.PreTokenBalances {

		// Convert string balances to big integers
		var pre, post big.Int
		_, ok1 := pre.SetString(tx.Meta.PreTokenBalances[i].UITokenAmount.Amount, 10)
		_, ok2 := post.SetString(tx.Meta.PostTokenBalances[i].UITokenAmount.Amount, 10)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("failed to convert token balances to integers")
		}

		idx := tx.Meta.PreTokenBalances[i].AccountIndex

		// Did the account's balance increase?
		if pre.Cmp(&post) == -1 { // pre < post
			if receiverFound {
				return nil, fmt.Errorf("unable to identify token receiver in txHash=%s", txHash)
			}
			receiverFound = true

			txData.Destination = tx.Transaction.Message.Accounts[idx].ToBase58()
			txData.Amount = big.NewInt(0)
			txData.Amount.Sub(&post, &pre)
		}

		// Did the account's balance decrease?
		if pre.Cmp(&post) == 1 { // pre > post
			if senderFound {
				return nil, fmt.Errorf("unable to identify token sender in txHash=%s", txHash)
			}
			senderFound = true

			txData.Source = tx.Transaction.Message.Accounts[idx].ToBase58()
			txData.Amount = big.NewInt(0)
			txData.Amount.Sub(&pre, &post)
		}
	}
	if !senderFound {
		return nil, fmt.Errorf("unable to identify participating addresses in txHash=%s", txHash)
	}

	return &txData, nil
}
