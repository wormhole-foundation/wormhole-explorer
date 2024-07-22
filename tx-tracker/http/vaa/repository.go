package vaa

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type VAARepository interface {
	GetVaa(ctx context.Context, id string) (*VaaDoc, error)
}

type RepositoryMongoDB struct {
	db                 *mongo.Database
	logger             *zap.Logger
	vaas               *mongo.Collection
	globalTransactions *mongo.Collection
}

type RepositoryPostreSQL struct {
	postreSQLClient *db.DB
}

func NewVaaRepositoryPostreSQL(postreSQLClient *db.DB) VAARepository {
	return &RepositoryPostreSQL{postreSQLClient: postreSQLClient}
}

type VaaDoc struct {
	ID     string `bson:"_id" json:"id"`
	Vaa    []byte `bson:"vaas" json:"vaa"`
	TxHash string `bson:"txHash" json:"txHash"`
	Digest string `bson:"digest" json:"digest"`
	Active bool   `bson:"active" json:"active"`
}

// NewMongoVaaRepository create a new VaaRepositoryMongoDB.
func NewMongoVaaRepository(db *mongo.Database, logger *zap.Logger) VAARepository {
	return &RepositoryMongoDB{db: db,
		logger:             logger.With(zap.String("module", "VaaRepository")),
		vaas:               db.Collection("vaas"),
		globalTransactions: db.Collection("globalTransactions"),
	}
}

func (r *RepositoryMongoDB) FindById(ctx context.Context, id string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	return &vaaDoc, err
}

func (r *RepositoryMongoDB) GetVaa(ctx context.Context, id string) (*VaaDoc, error) {
	return r.FindById(ctx, id)
}

func (r *RepositoryPostreSQL) GetVaa(ctx context.Context, id string) (*VaaDoc, error) {
	res := &VaaDoc{}
	err := r.postreSQLClient.SelectOne(
		ctx,
		res,
		"SELECT id,vaa_id,raw as vaas,active FROM wormhole.wh_attestation_vaas WHERE vaa_id = $1 and active = true", id)
	return res, err
}
