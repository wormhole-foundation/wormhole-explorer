// Package guardiansets contains historical guardian set information.
// TODO: HARDCODING for now. Let's get this from an ethereum node later.
package guardiansets

import (
	"github.com/certusone/wormhole/node/pkg/common"
	eth_common "github.com/ethereum/go-ethereum/common"
)

var ByIndexTest = []common.GuardianSet{gstest0}

func GetLatestTest() common.GuardianSet {
	return ByIndexTest[0]
}

// var gs0ValidUntil = time.Unix(1628599904, 0) // Tue Aug 10 2021 12:51:44 GMT+0000
var gstest0 = common.GuardianSet{
	Index: 0,
	Keys: []eth_common.Address{
		eth_common.HexToAddress("0x13947Bd48b18E53fdAeEe77F3473391aC727C638"), //
	},
}
