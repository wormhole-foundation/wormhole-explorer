package metrics

import (
	"math/big"
	"testing"

	"github.com/test-go/testify/assert"
)

// func Test_symbolLookup(t *testing.T) {

// 	symbol := symbolLookup("000000000000000000000000b1f66997a5760428d3a87d68b90bfe0ae64121cc")
// 	assert.Equal(t, "LUA", symbol)

// }

// func Test_symbolLookupSOL(t *testing.T) {

// 	symbol := symbolLookup("069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001")
// 	assert.Equal(t, "SOL", symbol)

// }

// //485493b637792cca16fe9d53fc4879c23dbf52cf6d9af4e61fe92df15c17c98d

// func Test_symbolLookupSOL2(t *testing.T) {

// 	symbol := symbolLookup("0001534f4c420000000000000000000000000000000000000000000000000000")
// 	assert.Equal(t, "SOL", symbol)

// }

// func Test_normalize1(t *testing.T) {

// 	n := normalizeAddress("00000000000000000000000055d398326f99059ff775485246999027b3197955")
// 	assert.Equal(t, "55d398326f99059ff775485246999027b3197955", n)

// }

// func Test_symbolLookupUSDT(t *testing.T) {

// 	//symbol := symbolLookup("00000000000000000000000055d398326f99059ff775485246999027b3197955")

// 	assert.Equal(t, "USDTbs", tokenMap["55d398326f99059ff775485246999027b3197955"])

// }

// // 55d398326f99059ff775485246999027b3197955

// func Test_base58encode(t *testing.T) {

// 	a := base58encode("c6fa7af3bedbad3a3d65f36aabc97431b1bbe4c2d2f6e0e47ca60203452f5d61")
// 	assert.Equal(t, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", a)

// 	symbol := symbolLookup(a)
// 	assert.Equal(t, "USDCso", symbol)

// }

func Test_formatAmount(t *testing.T) {

	r := formatAmount(big.NewInt(100000000))
	assert.Equal(t, "1.00000000", r)

	r = formatAmount(big.NewInt(16408113458008))
	assert.Equal(t, "164081.13458008", r)

}
