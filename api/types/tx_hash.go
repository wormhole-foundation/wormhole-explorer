package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
)

const solanaTxHashLen = 88
const ethMinTxHashLen = 64
const ethMaxTxHashLen = 66

// TxHash represents a transaction hash passed by query params.
type TxHash struct {
	hash     string
	isEth    bool
	isSolana bool
}

// ParseTxHash parses a transaction hash from a string.
//
// The transaction hash can be provided in different formats,
// depending on the blockchain it belongs to:
// * Solana: 64 bytes, encoded as base58.
// * All other chains: 32 bytes, encoded as hex.
//
// More cases could be added in the future as needed.
func ParseTxHash(value string) (*TxHash, error) {

	// Solana txHashes are 64 bytes long, encoded as base58.
	if len(value) == solanaTxHashLen {
		return parseSolanaTxHash(value)
	}

	// Ethereum txHashes are 32 bytes long, encoded as hex.
	// They can be prefixed with "0x" or "0X".
	if len(value) >= ethMinTxHashLen && len(value) <= ethMaxTxHashLen {
		return parseEthTxHash(value)
	}

	return nil, fmt.Errorf("invalid txHash length: %d", len(value))
}

func parseSolanaTxHash(value string) (*TxHash, error) {

	// Decode the string from base58 to binary
	bytes, err := base58.Decode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode txHash from base58: %w", err)
	}

	// Make sure we have the expected amount of bytes
	if len(bytes) != 64 {
		return nil, fmt.Errorf("solana txHash must be exactly 64 bytes, but got %d bytes", len(bytes))
	}

	// Populate the result struct and return
	result := TxHash{
		hash:     base58.Encode(bytes),
		isSolana: true,
	}
	return &result, nil
}

func parseEthTxHash(value string) (*TxHash, error) {

	// Trim any preceding "0x" to the address
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")

	// Decode the string from hex to binary
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode txHash from hex: %w", err)
	}

	// Make sure we have the expected amount of bytes
	if len(bytes) != 32 {
		return nil, fmt.Errorf("eth txHash must be exactly 32 bytes, but got %d bytes", len(bytes))
	}

	// Populate the result struct and return
	result := TxHash{
		hash:  hex.EncodeToString(bytes),
		isEth: true,
	}
	return &result, nil
}

func (h *TxHash) IsSolanaTxHash() bool {
	return h.isSolana
}

func (h *TxHash) IsEthTxHash() bool {
	return h.isEth
}

func (h *TxHash) String() string {
	return h.hash
}
