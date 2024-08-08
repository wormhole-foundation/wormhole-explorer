package chains

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"

	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type solanaTransactionSignature struct {
	BlockTime int64       `json:"blockTime"`
	Signature string      `json:"signature"`
	Err       interface{} `json:"err"`
}

type solanaGetTransactionResponse struct {
	BlockTime   int64  `json:"blockTime"`
	BlockNumber uint64 `json:"slot"`
	Meta        struct {
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
		Fee *uint64     `json:"fee"`
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
	timestamp     *time.Time
	notionalCache *notional.NotionalCache
	p2pNetwork    string
}

func (a *apiSolana) FetchSolanaTx(
	ctx context.Context,
	pool *pool.Pool,
	txHash string,
	metrics metrics.Metrics,
	logger *zap.Logger,
) (*TxDetail, error) {

	// get rpc sorted by score and priority.
	rpcs := pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	// Get the transaction from the Solana node API.
	var txDetail *TxDetail
	var err error
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		txDetail, err = a.fetchSolanaTx(ctx, rpc.Id, txHash)
		if txDetail != nil {
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDSolana), rpc.Description)
			break
		}
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDSolana), rpc.Description)
			logger.Debug("Failed to fetch transaction from Solana node", zap.String("url", rpc.Id), zap.Error(err))
		}
	}

	if txDetail != nil && txDetail.FeeDetail != nil && txDetail.FeeDetail.Fee != "" && a.p2pNetwork == domain.P2pMainNet {
		gasPrice, errGasPrice := GetGasTokenNotional(sdk.ChainIDSolana, a.notionalCache)
		if errGasPrice != nil {
			logger.Error("Failed to get gas price", zap.Error(errGasPrice), zap.String("chainId", sdk.ChainIDSolana.String()), zap.String("txHash", txHash))
		} else {
			txDetail.FeeDetail.GasTokenNotional = gasPrice.NotionalUsd.String()
			txDetail.FeeDetail.FeeUSD = gasPrice.NotionalUsd.Mul(decimal.RequireFromString(txDetail.FeeDetail.Fee)).String()
		}
	}

	return txDetail, err
}

func (a *apiSolana) fetchSolanaTx(
	ctx context.Context,
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
			err = client.CallContext(ctx, &sigs, "getSignaturesForAddress", base58.Encode(h))
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
		err = client.CallContext(ctx, &response, "getTransaction", nativeTxHash,
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

	respJson, _ := json.Marshal(response)

	// populate the response object
	txDetail := TxDetail{
		NativeTxHash: nativeTxHash,
		BlockNumber:  strconv.FormatUint(response.BlockNumber, 10),
		RpcResponse:  string(respJson),
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

	var feeDetail *FeeDetail
	if response.Meta.Fee != nil {
		feeDetail = &FeeDetail{
			RawFee: map[string]string{
				"fee": fmt.Sprintf("%d", *response.Meta.Fee),
			},
		}
		feeDetail.Fee = SolanaCalculateFee(*response.Meta.Fee)
		txDetail.FeeDetail = feeDetail
	}

	return &txDetail, nil
}

func SolanaCalculateFee(fee uint64) string {
	rawFee := decimal.NewFromUint64(fee)
	calculatedFee := rawFee.DivRound(decimal.NewFromInt(1e9), 9)
	return calculatedFee.String()
}
