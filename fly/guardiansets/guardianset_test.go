package guardiansets

import (
	_ "embed"
	"testing"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

//go:embed validVaa.bin
var validVaa []byte

// TestValidSignatures exercises the method `GuardianSetHistory.Validate()`
func TestValidSignatures(t *testing.T) {

	// unmarshal the binary encoding of the VAA into a high-level data structure
	var vaa sdk.VAA
	err := vaa.UnmarshalBinary(validVaa)
	if err != nil {
		t.Fatalf("Failed to unmarshal VAA: %v", err)
	}

	// assert that the signatures must be valid
	h := getMainnetGuardianSet()
	err = h.Verify(&vaa)
	if err != nil {
		t.Fatalf("Failed to verify VAA: %v", err)
	}

}

// TestInvalidSignatures exercises the method `GuardianSetHistory.Validate()`
func TestInvalidSignatures(t *testing.T) {

	// create an invalid VAA, binary encoded
	invalidVaa := make([]byte, len(validVaa))
	copy(invalidVaa, validVaa)
	invalidVaa[512] = 5 // changing a single byte in the signing body must render the signatures invalid

	// unmarshal the binary encoding of the VAA into a high-level data structure
	var vaa sdk.VAA
	err := vaa.UnmarshalBinary(invalidVaa)
	if err != nil {
		t.Fatalf("Failed to unmarshal VAA: %v", err)
	}

	// assert that the signatures must be invalid
	h := getMainnetGuardianSet()
	err = h.Verify(&vaa)
	if err == nil {
		t.Fatal("Expected signatures to be invalid")
	}

}
