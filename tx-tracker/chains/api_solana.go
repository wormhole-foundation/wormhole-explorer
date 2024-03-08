package chains

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mr-tron/base58"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
)

type solanaTransactionSignature struct {
	BlockTime int64       `json:"blockTime"`
	Signature string      `json:"signature"`
	Err       interface{} `json:"err"`
}

type solanaGetTransactionResponse struct {
	BlockTime int64 `json:"blockTime"`
	Meta      struct {
		InnerInstructions []struct {
			Instructions []struct {
				ParsedInstruction struct {
					Type_ string `json:"type"`
					Info  struct {
						Account     string `json:"account"`
						Amount      string `json:"amount"`
						Authority   string `json:"authority"`
						Destination string `json:"destination"`
						Source      string `json:"source"`
					} `json:"info"`
				} `json:"parsed"`
			} `json:"instructions"`
		} `json:"innerInstructions"`
		Err interface{} `json:"err"`
	} `json:"meta"`
	Transaction struct {
		Message struct {
			AccountKeys []struct {
				Pubkey string `json:"pubkey"`
				Signer bool   `json:"signer"`
			} `json:"accountKeys"`
		} `json:"message"`
		Signatures []string `json:"signatures"`
	} `json:"transaction"`
}

type getTransactionConfig struct {
	Encoding                       string `json:"encoding"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
}

type apiSolana struct {
	timestamp *time.Time
}

func (a *apiSolana) fetchSolanaTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Initialize RPC client
	client, err := rpcDialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// Decode txHash bytes
	// TODO: remove this when the fly fixes all txHash for Solana
	h, err := hex.DecodeString(txHash)
	if err != nil {
		h, err = base58.Decode(txHash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode from hex txHash=%s: %w", txHash, err)
		}
	}

	var sigs []solanaTransactionSignature
	nativeTxHash := txHash
	txHashType, err := types.ParseTxHash(txHash)
	isNotNativeTxHash := err != nil || !txHashType.IsSolanaTxHash()
	if isNotNativeTxHash {
		// Get transaction signatures for the given account
		{
			err = client.CallContext(ctx, rateLimiter, &sigs, "getSignaturesForAddress", base58.Encode(h))
			if err != nil {
				return nil, fmt.Errorf("failed to get signatures for account: %w (%+v)", err, err)
			}
			if len(sigs) == 0 {
				return nil, ErrTransactionNotFound
			}

			if len(sigs) == 1 {
				nativeTxHash = sigs[0].Signature
			} else {
				for _, sig := range sigs {

					if a.timestamp != nil && sig.BlockTime == a.timestamp.Unix() && sig.Err == nil {
						nativeTxHash = sig.Signature
						break
					}
				}
				if nativeTxHash == "" {
					return nil, fmt.Errorf("can't get signature, but found %d", len(sigs))
				}
			}
		}
	}

	// Fetch the portal token bridge transaction
	var response solanaGetTransactionResponse
	{
		err = client.CallContext(ctx, rateLimiter, &response, "getTransaction", nativeTxHash,
			getTransactionConfig{
				Encoding:                       "jsonParsed",
				MaxSupportedTransactionVersion: 0,
			})
		if err != nil {
			return nil, fmt.Errorf("failed to get tx by signature: %w", err)
		}
		if len(sigs) == 1 {
			if len(response.Meta.InnerInstructions) == 0 {
				return nil, fmt.Errorf("response.Meta.InnerInstructions is empty")
			}
			if len(response.Meta.InnerInstructions[0].Instructions) == 0 {
				return nil, fmt.Errorf("response.Meta.InnerInstructions[0].Instructions is empty")
			}
		}
	}

	// populate the response object
	txDetail := TxDetail{
		NativeTxHash: nativeTxHash,
	}

	// set sender/receiver
	for i := range response.Transaction.Message.AccountKeys {
		if response.Transaction.Message.AccountKeys[i].Signer {
			txDetail.From = response.Transaction.Message.AccountKeys[i].Pubkey
			// https://github.com/wormhole-foundation/wormhole-explorer/issues/1142
			// we get the first signer as the origintx from.
			break
		}
	}
	if txDetail.From == "" {
		return nil, fmt.Errorf("failed to find source account")
	}

	return &txDetail, nil
}
