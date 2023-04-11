package address

import "github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"

type AddressOverview struct {
	Vaas []*vaa.VaaDoc `json:"vaas"`
}
