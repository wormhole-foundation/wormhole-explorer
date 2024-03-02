package chains

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mr-tron/base58"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
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
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

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

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Initialize RPC client
		client, err := rpcDialContext(ctx, rpc.Id)
		if err != nil {
			logger.Error("failed to initialize RPC client", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
		}
		defer client.Close()

		if isNotNativeTxHash {
			// Get transaction signatures for the given
			err = client.CallContext(ctx, &sigs, "getSignaturesForAddress", base58.Encode(h))
			if err != nil {
				logger.Error("failed to get signatures for account", zap.Error(err), zap.String("rpc", rpc.Id))
				continue
			}

			if len(sigs) == 0 {
				logger.Error("no signatures found for account", zap.String("rpc", rpc.Id))
				continue
				//return nil, ErrTransactionNotFound
			}

			if len(sigs) == 1 {
				nativeTxHash = sigs[0].Signature
				fmt.Printf("rpc.Id = %s \n", rpc.Id)
				break

			} else {
				for _, sig := range sigs {

					if a.timestamp != nil && sig.BlockTime == a.timestamp.Unix() && sig.Err == nil {
						nativeTxHash = sig.Signature
						fmt.Printf("rpc.Id 2 = %s \n", rpc.Id)
						break
					}
				}
				if nativeTxHash == "" {
					continue
					//return nil, fmt.Errorf("can't get signature, but found %d", len(sigs))
				}
				fmt.Printf("rpc.Id 3 = %s \n", rpc.Id)
				break
			}
		}
	}

	// Check if we found a native tx hash
	if nativeTxHash == "" {
		return nil, fmt.Errorf("can't get signature, but found %d", len(sigs))
	}

	rpcs, err = getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	// Fetch the portal token bridge transaction
	var response *solanaGetTransactionResponse

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Initialize RPC client
		client, err := rpcDialContext(ctx, rpc.Id)
		if err != nil {
			logger.Error("failed to initialize RPC client", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
			//return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
		}
		defer client.Close()

		err = client.CallContext(ctx, &response, "getTransaction", nativeTxHash,
			getTransactionConfig{
				Encoding:                       "jsonParsed",
				MaxSupportedTransactionVersion: 0,
			})
		if err != nil {
			logger.Error("failed to get tx by signature", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
			//return nil, fmt.Errorf("failed to get tx by signature: %w", err)
		}
		if len(sigs) == 1 {
			if len(response.Meta.InnerInstructions) == 0 {
				logger.Error("response.Meta.InnerInstructions is empty", zap.String("rpc", rpc.Id))
				continue
				//return nil, fmt.Errorf("response.Meta.InnerInstructions is empty")
			}
			if len(response.Meta.InnerInstructions[0].Instructions) == 0 {
				logger.Error("response.Meta.InnerInstructions[0].Instructions is empty", zap.String("rpc", rpc.Id))
				continue
				//return nil, fmt.Errorf("response.Meta.InnerInstructions[0].Instructions is empty")
			}
			fmt.Printf("rpc.Id 4 = %s \n", rpc.Id)
			break
		}
	}

	// Check if we found a response
	if response == nil {
		return nil, ErrTransactionNotFound
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
