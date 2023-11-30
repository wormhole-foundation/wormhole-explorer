package types

import (
	"fmt"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Address struct {
	address vaa.Address
}

func BytesToAddress(b []byte) (*Address, error) {

	var a Address

	if len(b) != len(a.address) {
		return nil, fmt.Errorf("expected byte slice to have len=%d, but got %d instead", len(a.address), len(b))
	}

	copy(a.address[:], b)

	return &a, nil
}

// StringToAddress converts a hex-encoded address string into an *Address.
func StringToAddress(s string, acceptSolanaFormat bool) (*Address, error) {

	// Attempt to parse a Solana address (i.e.: 32 bytes encoded as base58).
	// If it fails, fall back to parsing a regular Wormhole address.
	if acceptSolanaFormat {

		sig, err := solana.PublicKeyFromBase58(s)
		if err == nil {
			// This step is not expected to fail, since Solana and Wormhole addresses have the same size.
			emitter, err := BytesToAddress(sig[:])
			if err == nil {
				return emitter, nil
			}
		}
	}

	// Attempt to parse a regular Wormhole address (i.e.: 32 bytes encoded as hex).
	a, err := vaa.StringToAddress(s)
	if err != nil {
		return nil, err
	}

	return &Address{address: a}, nil
}

// Hex returns the full 32-byte address, encoded as hex.
func (addr *Address) Hex() string {
	return addr.address.String()
}

// ShortHex returns a hex-encoded address that is usually shorted than Hex().
//
// If the full address returned by Hex() is prefixed 12 bytes set to zero,
// this function will trim those bytes.
//
// The reason we need this function is that a few database collections
// (governorConfig, governorStatus, heartbeats) store guardian addresses
// as 40 hex digits instead of the full 64-digit hex representation.
// When performing lookups over those collections, this function
// can perform the conversion.
func (addr *Address) ShortHex() string {

	full := addr.Hex()

	if len(full) == 64 && strings.HasPrefix(full, "000000000000000000000000") {
		return full[24:]
	}

	return full
}

// Copy returns a deep copy of the address.
func (addr *Address) Copy() *Address {

	var tmp vaa.Address
	copy(tmp[:], addr.address[:])

	return &Address{address: tmp}
}
