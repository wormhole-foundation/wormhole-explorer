package storage

import (
	"context"
	"time"

	"errors"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type GovernorStatusRepository interface {
	FindNodeGovernorVaaByNodeAddress(ctx context.Context, nodeAddress string) ([]NodeGovernorVaa, error)
	FindNodeGovernorVaaByVaaID(ctx context.Context, vaaID string) ([]NodeGovernorVaa, error)
	FindNodeGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]NodeGovernorVaa, error)
	FindGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]GovernorVaa, error)
	UpdateGovernorStatus(ctx context.Context, nodeGovernorVaaDocToInsert []NodeGovernorVaa,
		nodeGovernorVaaDocToDelete []string, governorVaasToInsert []GovernorVaa,
		governorVaaIdsToDelete []string) error
}

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

type NodeGovernorVaa struct {
	ID          string     `bson:"_id" db:"guardian_address"` //TODO check if this is correct
	NodeName    string     `bson:"nodeName" db:"guardian_name"`
	NodeAddress string     `bson:"nodeAddress" db:"guardian_address"`
	VaaID       string     `bson:"vaaId" db:"vaa_id"`
	CreatedAt   *time.Time `bson:"-" db:"created_at"`
	UpdatedAt   *time.Time `bson:"-" db:"updated_at"`
}

type GovernorVaa struct {
	ID             string      `bson:"_id" db:"id"`
	ChainID        sdk.ChainID `bson:"chainId" db:"chain_id"`
	EmitterAddress string      `bson:"emitterAddress" db:"emitter_address"`
	Sequence       string      `bson:"sequence" db:"sequence"`
	TxHash         string      `bson:"txHash" db:"tx_hash"`
	ReleaseTime    time.Time   `bson:"releaseTime" db:"release_time"`
	Amount         Uint64      `bson:"amount" db:"notional_value"`
	CreatedAt      *time.Time  `bson:"-" db:"created_at"`
	UpdatedAt      *time.Time  `bson:"-" db:"updated_at"`
}

type GovernorConfigChain struct {
	GovernorConfigID   string    `db:"governor_config_id"`
	ChainID            uint16    `db:"chain_id"`
	NotionalLimit      uint64    `db:"notional_limit"`
	BigTransactionSize uint64    `db:"big_transaction_size"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
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

type AttestationVaa struct {
	ID             string      `db:"id"`
	VaaID          string      `db:"vaa_id"`
	Version        uint8       `db:"version"`
	EmitterChain   sdk.ChainID `db:"emitter_chain_id"`
	EmitterAddress string      `db:"emitter_address"`
	Sequence       Uint64      `db:"sequence"`
	GuardianSetIdx uint32      `db:"guardian_set_index"`
	Raw            []byte      `db:"raw"`
	Timestamp      time.Time   `db:"timestamp"`
	Active         bool        `db:"active"`
	IsDuplicated   bool        `db:"is_duplicated"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      *time.Time  `db:"updated_at"`
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
