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
	logger          *zap.Logger
}

func NewVaaRepositoryPostreSQL(postreSQLClient *db.DB, logger *zap.Logger) VAARepository {
	return &RepositoryPostreSQL{
		postreSQLClient: postreSQLClient,
		logger:          logger,
	}
}

type VaaDoc struct {
	VaaID  string `bson:"vaa_id" json:"vaa_id"`
	ID     string `bson:"_id" json:"id"`
	Vaa    []byte `bson:"vaas" json:"vaa"`
	TxHash string `bson:"txHash" json:"txHash"`
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
		"SELECT id,vaa_id,raw as vaas,active FROM wormholescan.wh_attestation_vaas WHERE vaa_id = $1 and active = true",
		id)

	if err != nil {
		// fallback: in case the vaa is not found in the attestation_vaas table, try to find it in the operation_transactions table to grab the digest and tx_hash
		r.logger.Debug("Failed to get vaa from wh_attestation_vaas table", zap.Error(err), zap.String("vaa_id", id))
		err = r.postreSQLClient.SelectOne(
			ctx,
			res,
			"SELECT attestation_vaas_id as id, vaa_id, tx_hash FROM wormholescan.wh_operation_transactions WHERE vaa_id = $1 LIMIT 1", // LIMIT 1 is due to wormchain transactions which have 2 txs.
			id)
	}
	return res, err
}
