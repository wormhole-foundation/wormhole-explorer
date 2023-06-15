package chains

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mr-tron/base58"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

type solanaTransactionSignature struct {
	Signature string `json:"signature"`
}

type solanaGetTransactionResponse struct {
	BlockTime   int64                 `json:"blockTime"`
	Meta        solanaTransactionMeta `json:"meta"`
	Transaction solanaTransaction     `json:"transaction"`
}

type solanaTransactionMeta struct {
	InnerInstructions []solanaInnerInstruction `json:"innerInstructions"`
	Err               []interface{}            `json:"err"`
}

type solanaInnerInstruction struct {
	Instructions []solanaInstruction `json:"instructions"`
}

type solanaInstruction struct {
	ParsedInstruction solanaParsedInstruction `json:"parsed"`
}

type solanaParsedInstruction struct {
	Type_ string                      `json:"type"`
	Info  solanaParsedInstructionInfo `json:"info"`
}

type solanaParsedInstructionInfo struct {
	Account     string `json:"account"`
	Amount      string `json:"amount"`
	Authority   string `json:"authority"`
	Destination string `json:"destination"`
	Source      string `json:"source"`
}

type solanaTransaction struct {
	Message    solanaTransactionMessage `json:"message"`
	Signatures []string                 `json:"signatures"`
}

type solanaTransactionMessage struct {
	AccountKeys []solanaAccountKey `json:"accountKeys"`
}

type solanaAccountKey struct {
	Pubkey string `json:"pubkey"`
	Signer bool   `json:"signer"`
}

func fetchSolanaTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	txHash string,
) (*TxDetail, error) {

	// Initialize RPC client
	client, err := rpc.DialContext(ctx, cfg.SolanaBaseUrl)
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

	// Get transaction signatures for the given account
	var sigs []solanaTransactionSignature
	err = client.CallContext(ctx, &sigs, "getSignaturesForAddress", base58.Encode(h))
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures for account: %w (%+v)", err, err)
	}
	if len(sigs) == 0 {
		return nil, ErrTransactionNotFound
	}
	if len(sigs) > 1 {
		return nil, fmt.Errorf("expected exactly one signature, but found %d", len(sigs))
	}

	// Fetch the portal token bridge transaction
	var response solanaGetTransactionResponse
	err = client.CallContext(ctx, &response, "getTransaction", sigs[0].Signature, "jsonParsed")
	if err != nil {
		return nil, fmt.Errorf("failed to get tx by signature: %w", err)
	}
	if len(response.Meta.InnerInstructions) == 0 {
		return nil, fmt.Errorf("response.Meta.InnerInstructions is empty")
	}
	if len(response.Meta.InnerInstructions[0].Instructions) == 0 {
		return nil, fmt.Errorf("response.Meta.InnerInstructions[0].Instructions is empty")
	}

	// populate the response object
	txDetail := TxDetail{
		Timestamp:    time.Unix(response.BlockTime, 0).UTC(),
		NativeTxHash: sigs[0].Signature,
	}

	// set sender/receiver
	for i := range response.Transaction.Message.AccountKeys {
		if response.Transaction.Message.AccountKeys[i].Signer {
			txDetail.From = response.Transaction.Message.AccountKeys[i].Pubkey
		}
	}
	if txDetail.From == "" {
		return nil, fmt.Errorf("failed to find source account")
	}

	return &txDetail, nil
}
