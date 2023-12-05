package types

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
)

const (
	suiMinTxHashLen      = 43
	suiMaxTxHashLen      = 44
	algorandTxHashLen    = 52
	wormholeMinTxHashLen = 64
	wormholeMaxTxHashLen = 66
	solanaMinTxHashLen   = 87
	solanaMaxTxHashLen   = 88
)

// TxHash represents a transaction hash passed by query params.
type TxHash struct {
	hash       string
	isWormhole bool
	isSolana   bool
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
	if len(value) >= solanaMinTxHashLen && len(value) <= solanaMaxTxHashLen {
		return parseSolanaTxHash(value)
	}

	// Algorand txHashes are 32 bytes long, encoded as base32.
	if len(value) == algorandTxHashLen {
		return parseAlgorandTxHash(value)
	}

	// Sui txHashes are 32 bytes long, encoded as base32.
	if len(value) >= suiMinTxHashLen && len(value) <= suiMaxTxHashLen {
		return parseSuiTxHash(value)
	}

	// Wormhole txHashes are 32 bytes long, encoded as hex.
	// Optionally, they can be prefixed with "0x" or "0X".
	if len(value) >= wormholeMinTxHashLen && len(value) <= wormholeMaxTxHashLen {
		return parseWormholeTxHash(value)
	}

	return nil, fmt.Errorf("invalid txHash length: %d", len(value))
}

func parseSolanaTxHash(value string) (*TxHash, error) {

	// Decode the string from base58 to binary
	bytes, err := base58.Decode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode solana txHash from base58: %w", err)
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

func parseAlgorandTxHash(value string) (*TxHash, error) {
	// Decode the string from base32 to binary
	bytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode algorand txHash from base32: %w", err)
	}

	// Make sure we have the expected amount of bytes
	if len(bytes) != 32 {
		return nil, fmt.Errorf("algorand txHash must be exactly 32 bytes, but got %d bytes", len(bytes))
	}

	// Populate the result struct and return
	result := TxHash{
		hash:       base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes),
		isWormhole: true,
	}
	return &result, nil
}

func parseSuiTxHash(value string) (*TxHash, error) {

	// Decode the string from base58 to binary
	bytes, err := base58.Decode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sui txHash from base58: %w", err)
	}

	// Make sure we have the expected amount of bytes
	if len(bytes) != 32 {
		return nil, fmt.Errorf("sui txHash must be exactly 32 bytes, but got %d bytes", len(bytes))
	}

	// Populate the result struct and return
	result := TxHash{
		hash:       base58.Encode(bytes),
		isWormhole: true,
	}
	return &result, nil
}

func parseWormholeTxHash(value string) (*TxHash, error) {

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
		return nil, fmt.Errorf("wormhole txHash must be exactly 32 bytes, but got %d bytes", len(bytes))
	}

	// Populate the result struct and return
	result := TxHash{
		hash:       hex.EncodeToString(bytes),
		isWormhole: true,
	}
	return &result, nil
}

func (h *TxHash) IsSolanaTxHash() bool {
	return h.isSolana
}

func (h *TxHash) IsWormholeTxHash() bool {
	return h.isWormhole
}

func (h *TxHash) String() string {
	return h.hash
}
