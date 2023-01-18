// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/certusone/wormhole/node/pkg/vaa"
)

// ObservationDoc represent an observation document.
type ObservationDoc struct {
	ID           string      `bson:"_id" json:"id"`
	Version      uint8       `bson:"version" json:"version"`
	EmitterChain vaa.ChainID `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr  string      `bson:"emitterAddr" json:"emitterAddr"`
	Sequence     string      `bson:"sequence" json:"sequence"`
	Hash         []byte      `bson:"hash" json:"hash"`
	TxHash       []byte      `bson:"txHash" json:"txHash"`
	GuardianAddr string      `bson:"guardianAddr" json:"guardianAddr"`
	Signature    []byte      `bson:"signature" json:"signature"`
	UpdatedAt    *time.Time  `bson:"updatedAt" json:"updatedAt"`
	IndexedAt    *time.Time  `bson:"indexedAt" json:"indexedAt"`
}

// MarshalJSON interface implementation for ObservationDoc.
func (o *ObservationDoc) MarshalJSON() ([]byte, error) {
	sequence, err := strconv.ParseUint(o.Sequence, 10, 64)
	if err != nil {
		return []byte{}, err
	}

	type Alias ObservationDoc
	return json.Marshal(&struct {
		Sequence uint64 `json:"sequence"`
		*Alias
	}{
		Sequence: sequence,
		Alias:    (*Alias)(o),
	})
}
