package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// repository errors
var ErrNotFound = errors.New("NOT FOUND")

// Repository definitions.
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		vaaParserFunctions *mongo.Collection
	}
}

// NewRepository create a new respository.
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		vaaParserFunctions *mongo.Collection
	}{
		vaaParserFunctions: db.Collection("vaaParserFunctions"),
	}}
}

// GetVaaParserFunction get a vaa parser function by chainID and address.
func (r *Repository) GetVaaParserFunction(ctx context.Context, chainID vaa.ChainID, address string) (*VaaParserFunctions, error) {
	filter := bson.D{bson.E{"emitterChain", chainID}, bson.E{"emitterAddress", address}}
	var vpf VaaParserFunctions
	err := r.collections.vaaParserFunctions.FindOne(ctx, filter).Decode(&vpf)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		r.log.Error("failed execute FindOne command to get vaaParserFunctions",
			zap.Error(err), zap.Uint16("chainID", uint16(chainID)), zap.String("address", address))
		return nil, err
	}
	return &vpf, nil
}
