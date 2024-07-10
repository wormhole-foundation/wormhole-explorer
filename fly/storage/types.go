package storage

import (
	"context"
	"errors"
	"strconv"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Storage is a storage interface.
type Storage interface {
	UpsertObservation(ctx context.Context, o *gossipv1.SignedObservation, saveTxHash bool) error
	UpsertVAA(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error
	ReplaceVaaTxHash(ctx context.Context, vaaID string, oldTxHash string, newTxHash string) error // TODO: evaluate backfiller process.
	UpsertHeartbeat(hb *gossipv1.Heartbeat) error
	UpsertGovernorConfig(govC *gossipv1.SignedChainGovernorConfig) error
	UpsertGovernorStatus(govS *gossipv1.SignedChainGovernorStatus) error
	FindVaaByID(ctx context.Context, vaaID string) (*VaaUpdate, error) // TODO change VaaUpdate
	FindVaaByChainID(ctx context.Context, chainID vaa.ChainID, page int64, pageSize int64) ([]*VaaUpdate, error)
	UpsertDuplicateVaa(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error
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
