package storage

import (
	"time"

	"errors"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
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

type NodeGovernorVaaDoc struct {
	ID          string `bson:"_id"` //--> nodeAddress-vaaId
	NodeName    string `bson:"nodeName"`
	NodeAddress string `bson:"nodeAddress"`
	VaaID       string `bson:"vaaId"`
}

type GovernorVaaDoc struct {
	ID             string      `bson:"_id"` // --> vaaId
	ChainID        sdk.ChainID `bson:"chainId"`
	EmitterAddress string      `bson:"emitterAddress"`
	Sequence       string      `bson:"sequence"`
	TxHash         string      `bson:"txHash"` //Message // governorVaa // Global Transactions // tx-tracker
	ReleaseTime    time.Time   `bson:"releaseTime"`
	Amount         Uint64      `bson:"amount"`
	Status         string      `bson:"status"` //vaa //
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
	vaa, err := sdk.Unmarshal(v.Vaa)
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

type Uint64 uint64

func (u Uint64) MarshalBSONValue() (bsontype.Type, []byte, error) {
	ui64Str := strconv.FormatUint(uint64(u), 10)
	d128, err := primitive.ParseDecimal128(ui64Str)
	return bsontype.Decimal128, bsoncore.AppendDecimal128(nil, d128), err
}

func (u *Uint64) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	d128, _, ok := bsoncore.ReadDecimal128(b)
	if !ok {
		return errors.New("Uint64 UnmarshalBSONValue error")
	}

	ui64, err := strconv.ParseUint(d128.String(), 10, 64)
	if err != nil {
		return err
	}

	*u = Uint64(ui64)
	return nil
}
