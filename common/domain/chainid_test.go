package domain

import (
	"testing"

	"github.com/test-go/testify/assert"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func TestEncodeTrxHashByChainID(t *testing.T) {
	var tests = []struct {
		chainID sdk.ChainID
		txHash  []byte
		want    string
		err     error
	}{
		{
			chainID: sdk.ChainIDSolana,
			txHash:  []byte{0x23, 0xac, 0x49, 0x94, 0x37, 0xa8, 0xe6, 0x53, 0x3b, 0x79, 0x0d, 0x55, 0x78, 0xaf, 0x5d, 0x39, 0xb3, 0x49, 0x88, 0x31, 0x88, 0xec, 0xa5, 0x35, 0xb9, 0x57, 0xd8, 0x2a, 0x0e, 0x77, 0xeb, 0x03},
			want:    "3QFeCHsG9WDXzMozWyck8RUxmw59jyj7MPnQd4w2mbDL",
			err:     nil,
		},
		{
			chainID: sdk.ChainIDAlgorand,
			txHash:  []byte{0xd3, 0x45, 0x59, 0x5e, 0x2a, 0x0f, 0xab, 0x5c, 0xde, 0x71, 0x20, 0xb6, 0xbe, 0xb6, 0xee, 0x0b, 0xb9, 0x4b, 0x57, 0x8a, 0xa5, 0x69, 0x95, 0x2d, 0x00, 0x0c, 0xe8, 0xbf, 0xef, 0x03, 0x2d, 0x22},
			want:    "2NCVSXRKB6VVZXTREC3L5NXOBO4UWV4KUVUZKLIABTUL73YDFURA",
			err:     nil,
		},
		{
			chainID: sdk.ChainIDEthereum,
			txHash:  []byte{0xb9, 0x11, 0xcb, 0xfb, 0x0e, 0x42, 0xc5, 0x04, 0x77, 0x2b, 0xe9, 0x16, 0xbb, 0xeb, 0x8a, 0x46, 0xfc, 0xe7, 0x2b, 0xe5, 0xc6, 0x11, 0x28, 0xe7, 0x12, 0x93, 0x68, 0x26, 0x32, 0x88, 0xcc, 0x7d},
			want:    "b911cbfb0e42c504772be916bbeb8a46fce72be5c61128e7129368263288cc7d",
			err:     nil,
		},
		{
			chainID: sdk.ChainIDNear,
			txHash:  []byte{0x02, 0xde, 0x67, 0xd0, 0x15, 0x34, 0x02, 0x1c, 0x0e, 0x5b, 0x17, 0x68, 0x6e, 0x1e, 0x70, 0xd4, 0x79, 0x39, 0x6d, 0xa2, 0x9d, 0x1e, 0xbc, 0xe4, 0x9a, 0x4c, 0xad, 0xda, 0x4b, 0xca, 0xa3, 0x2b},
			want:    "CCWhFHoDg5eycFJC7EHbYXnNdXW1ed8tjdNHCLbYZEa",
			err:     nil,
		},
		{
			chainID: sdk.ChainIDSui,
			txHash:  []byte{0x29, 0xf4, 0xe6, 0xd8, 0xe0, 0xbf, 0x65, 0x21, 0xe5, 0xf3, 0x30, 0x28, 0x73, 0xa1, 0xf0, 0x08, 0x65, 0xb7, 0xcf, 0xe0, 0x48, 0x36, 0x73, 0x4d, 0x74, 0xed, 0x8c, 0x99, 0x6e, 0x7a, 0x07, 0x86},
			want:    "3pnJrxdjJeDUSvAquDiidApuRLXp5jATdLPyLhjrJsv5",
			err:     nil,
		},
	}

	for _, test := range tests {
		got, err := EncodeTrxHashByChainID(test.chainID, test.txHash)
		assert.Equal(t, test.want, got, "EncodeTrxHashByChainID() = %v, want %v", got, test.want)
		assert.Equal(t, test.err, err, "EncodeTrxHashByChainID() = %v, want %v", err, test.err)
	}
}

// TestTranslateEmitterAddress contains a test harness for the `TranslateEmitterAddress` function.
func TestTranslateEmitterAddress(t *testing.T) {

	// A table defining the test cases
	tcs := []struct {
		emitterChain   sdk.ChainID
		emitterAddress string
		want           string
		err            error
	}{
		{
			// Solana - Token Bridge emitter
			emitterChain:   sdk.ChainIDSolana,
			emitterAddress: "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
			want:           "Gv1KWf8DT1jKv5pKBmGaTmVszqa56Xn8YGx2Pg7i7qAk",
		},
		{
			// Ethereum - Token Bridge emitter
			emitterChain:   sdk.ChainIDEthereum,
			emitterAddress: "0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585",
			want:           "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
		},
		{
			// Terra - Token Bridge emitter
			emitterChain:   sdk.ChainIDTerra,
			emitterAddress: "0000000000000000000000007cf7b764e38a0a5e967972c1df77d432510564e2",
			want:           "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
		},
		{
			// BSC - Token Bridge emitter
			emitterChain:   sdk.ChainIDBSC,
			emitterAddress: "000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7",
			want:           "0xb6f6d86a8f9879a9c87f643768d9efc38c1da6e7",
		},
		{
			// Polygon - Token Bridge emitter
			emitterChain:   sdk.ChainIDPolygon,
			emitterAddress: "0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde",
			want:           "0x5a58505a96d1dbf8df91cb21b54419fc36e93fde",
		},
		{
			// Avalanche - Token Bridge emitter
			emitterChain:   sdk.ChainIDAvalanche,
			emitterAddress: "0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052",
			want:           "0x0e082f06ff657d94310cb8ce8b0d9a04541d8052",
		},
	}

	// For each test case
	for i := range tcs {
		tc := &tcs[i]

		// Make sure that the function returns the expected value
		emitterNativeAddress, err := TranslateEmitterAddress(tc.emitterChain, tc.emitterAddress)
		if err != tc.err {
			t.Fatalf("TranslateEmitterAddress(%s,%s)=%v, want=%v", tc.emitterChain.String(), tc.emitterAddress, err, tc.err)
		}
		if emitterNativeAddress != tc.want {
			t.Fatalf(`TranslateEmitterAddress(%s,%s)="%s", want="%s"`, tc.emitterChain.String(), tc.emitterAddress, emitterNativeAddress, tc.want)
		}
	}
}
