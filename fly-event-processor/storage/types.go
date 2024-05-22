package storage

import (
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VaaDoc represents a VAA document.
type VaaDoc struct {
	ID               string      `bson:"_id"`
	Version          uint8       `bson:"version"`
	EmitterChain     sdk.ChainID `bson:"emitterChain"`
	EmitterAddr      string      `bson:"emitterAddr"`
	Sequence         string      `bson:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas"`
	TxHash           string      `bson:"txHash,omitempty"`
	OriginTxHash     *string     `bson:"_originTxHash,omitempty"` //this is temporary field for fix enconding txHash
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
	Digest           string      `bson:"digest"`
	IsDuplicated     bool        `bson:"isDuplicated"`
	DuplicatedFixed  bool        `bson:"duplicatedFixed"`
}

// DuplicateVaaDoc represents a duplicate VAA document.
type DuplicateVaaDoc struct {
	ID               string      `bson:"_id"`
	VaaID            string      `bson:"vaaId"`
	Version          uint8       `bson:"version"`
	EmitterChain     sdk.ChainID `bson:"emitterChain"`
	EmitterAddr      string      `bson:"emitterAddr"`
	Sequence         string      `bson:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas"`
	Digest           string      `bson:"digest"`
	ConsistencyLevel uint8       `bson:"consistencyLevel"`
	TxHash           string      `bson:"txHash,omitempty"`
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
}

func (d *DuplicateVaaDoc) ToVaaDoc(duplicatedFixed bool) *VaaDoc {
	return &VaaDoc{
		ID:               d.VaaID,
		Version:          d.Version,
		EmitterChain:     d.EmitterChain,
		EmitterAddr:      d.EmitterAddr,
		Sequence:         d.Sequence,
		GuardianSetIndex: d.GuardianSetIndex,
		Vaa:              d.Vaa,
		Digest:           d.Digest,
		TxHash:           d.TxHash,
		OriginTxHash:     nil,
		Timestamp:        d.Timestamp,
		UpdatedAt:        d.UpdatedAt,
		DuplicatedFixed:  duplicatedFixed,
		IsDuplicated:     true,
	}
}

func (v *VaaDoc) ToDuplicateVaaDoc() (*DuplicateVaaDoc, error) {
	vaa, err := vaa.Unmarshal(v.Vaa)
	if err != nil {
		return nil, err
	}

	uniqueId := domain.CreateUniqueVaaID(vaa)
	return &DuplicateVaaDoc{
		ID:               uniqueId,
		VaaID:            v.ID,
		Version:          v.Version,
		EmitterChain:     v.EmitterChain,
		EmitterAddr:      v.EmitterAddr,
		Sequence:         v.Sequence,
		GuardianSetIndex: v.GuardianSetIndex,
		Vaa:              v.Vaa,
		Digest:           v.Digest,
		TxHash:           v.TxHash,
		ConsistencyLevel: vaa.ConsistencyLevel,
		Timestamp:        v.Timestamp,
		UpdatedAt:        v.UpdatedAt,
	}, nil
}
