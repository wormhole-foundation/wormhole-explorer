package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	vaaRepo "github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// DestinationTx representa a destination transaction.
type DestinationTx struct {
	ChainID     sdk.ChainID `bson:"chainId"`
	Status      string      `bson:"status"`
	Method      string      `bson:"method"`
	TxHash      string      `bson:"txHash"`
	From        string      `bson:"from"`
	To          string      `bson:"to"`
	BlockNumber string      `bson:"blockNumber"`
	Timestamp   *time.Time  `bson:"timestamp"`
	FeeDetail   *FeeDetail  `bson:"feeDetail"`
	UpdatedAt   *time.Time  `bson:"updatedAt"`
}

type FeeDetail struct {
	Fee              string            `bson:"fee" json:"fee"`
	RawFee           map[string]string `bson:"rawFee" json:"rawFee"`
	GasTokenNotional string            `bson:"gasTokenNotional" json:"gasTokenNotional"`
	FeeUSD           string            `bson:"feeUSD" json:"feeUSD"`
}

// TargetTxUpdate represents a transaction document.
type TargetTxUpdate struct {
	ID          string         `bson:"digest"`
	VaaID       string         `bson:"_id"`
	Destination *DestinationTx `bson:"destinationTx"`
	Source      string         `bson:"-"`
	TrackID     string         `bson:"-"`
}

// Repository exposes operations over the `globalTransactions` collection.
type MongoRepository struct {
	metrics            metrics.Metrics
	logger             *zap.Logger
	globalTransactions *mongo.Collection
	vaaIdTxHash        *mongo.Collection
	vaaRepository      *vaaRepo.RepositoryMongoDB
}

// New creates a new repository.
func NewMongoRepository(logger *zap.Logger, db *mongo.Database, vaaRepository *vaaRepo.RepositoryMongoDB,
	metrics metrics.Metrics) *MongoRepository {
	r := MongoRepository{
		metrics:            metrics,
		logger:             logger,
		globalTransactions: db.Collection("globalTransactions"),
		vaaIdTxHash:        db.Collection("vaaIdTxHash"),
		vaaRepository:      vaaRepository,
	}

	return &r
}

// UpsertOriginTxParams is a struct that contains the parameters for the upsertDocument method.
type UpsertOriginTxParams struct {
	VaaId     string // {chain/address/sequence}
	Id        string // digest
	TxType    string
	Source    string
	TrackID   string
	ChainId   sdk.ChainID
	TxDetail  *chains.TxDetail
	TxStatus  domain.SourceTxStatus
	Timestamp *time.Time
	Processed bool
}

func createChangesDoc(source, _type string, timestamp *time.Time) bson.D {
	return bson.D{
		{
			Key: "changes",
			Value: bson.D{
				{Key: "type", Value: _type},
				{Key: "source", Value: source},
				{Key: "timestamp", Value: timestamp},
			},
		},
	}
}

// UpsertOriginTx upserts a source transaction document.
func (r *MongoRepository) UpsertOriginTx(ctx context.Context, originTx, _ *UpsertOriginTxParams) error {

	now := time.Now()

	fields := bson.D{
		{Key: "chainId", Value: originTx.ChainId},
		{Key: "status", Value: originTx.TxStatus},
		{Key: "updatedAt", Value: now},
		{Key: "processed", Value: originTx.Processed},
	}

	if originTx.TxDetail != nil {
		fields = append(fields, primitive.E{Key: "nativeTxHash", Value: originTx.TxDetail.NativeTxHash})
		fields = append(fields, primitive.E{Key: "from", Value: originTx.TxDetail.From})
		if originTx.TxDetail.Attribute != nil {
			fields = append(fields, primitive.E{Key: "attribute", Value: originTx.TxDetail.Attribute})
		}
		if originTx.TxDetail.FeeDetail != nil {
			fields = append(fields, primitive.E{Key: "feeDetail", Value: originTx.TxDetail.FeeDetail})
		}
	}

	if originTx.Timestamp != nil {
		fields = append(fields, primitive.E{Key: "timestamp", Value: originTx.Timestamp})
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "originTx",
					Value: fields,
				},
			},
		},
		{
			Key:   "$push",
			Value: createChangesDoc(originTx.TrackID, "originTx", &now),
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.globalTransactions.UpdateByID(ctx, originTx.VaaId, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert source tx information: %w", err)
	}

	r.metrics.IncGlobalTxSourceInserted(uint16(originTx.ChainId))
	return nil
}

// AlreadyProcessed returns true if the given VAA ID has already been processed.
func (r *MongoRepository) AlreadyProcessed(ctx context.Context, vaaId string, _ string) (bool, error) {
	result := r.
		globalTransactions.
		FindOne(ctx, bson.D{
			{Key: "_id", Value: vaaId},
			{Key: "originTx", Value: bson.D{{Key: "$exists", Value: true}}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "originTx.processed", Value: true}},
				bson.D{{Key: "originTx.processed", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		})
	//  The originTx.processed will be true if the vaa was processed successfully.
	//  If exists and error getting the transactions from the rpcs, a partial originTx will save in the db and
	//  the originTx.processed will be false.

	var tx GlobalTransaction
	err := result.Decode(&tx)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to decode already processed VAA id: %w", err)
	} else {
		return true, nil
	}
}

type GlobalTransaction struct {
	Id   string       `bson:"_id"`
	Vaas []vaa.VaaDoc `bson:"vaas"`
}

// VaaIdTxHash represents a vaaIdTxHash document.
type VaaIdTxHash struct {
	TxHash string `bson:"txHash"`
}

func (r *MongoRepository) GetVaaIdTxHash(ctx context.Context, vaaID string, vaaDigest string) (*VaaIdTxHash, error) {
	var v VaaIdTxHash
	err := r.vaaIdTxHash.FindOne(ctx, bson.M{"_id": vaaID}).Decode(&v)
	return &v, err
}

func (r *MongoRepository) UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error {
	update := bson.M{
		"$set":  globalTx,
		"$push": createChangesDoc(globalTx.TrackID, "destinationTx", globalTx.Destination.UpdatedAt),
	}

	_, err := r.globalTransactions.UpdateByID(ctx, globalTx.VaaID, update, options.Update().SetUpsert(true))
	if err != nil {
		r.logger.Error("Error inserting target tx in global transaction", zap.Error(err))
		return err
	}
	r.metrics.IncGlobalTxDestinationTxInserted(uint16(globalTx.Destination.ChainID))
	return err
}

// GetTxStatus returns the status of the transaction with the given VAA ID.
func (r *MongoRepository) GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error) {

	result := r.
		globalTransactions.
		FindOne(ctx, bson.D{
			{Key: "_id", Value: targetTxUpdate.VaaID},
			{Key: "destinationTx", Value: bson.D{{Key: "$exists", Value: true}}},
		})

	var tx TargetTxUpdate
	err := result.Decode(&tx)
	if err == nil {
		return tx.Destination.TxHash, nil
	} else if err != mongo.ErrNoDocuments {
		return "", fmt.Errorf("failed to decode already processed VAA id: %w", err)
	} else {
		return "", nil
	}
}

// SourceTxDoc represents a source transaction document.
type SourceTxDoc struct {
	ID       string `bson:"_id"`
	OriginTx *struct {
		ChainID      int    `bson:"chainId"`
		Status       string `bson:"status"`
		Processed    bool   `bson:"processed"`
		NativeTxHash string `bson:"nativeTxHash"`
		From         string `bson:"from"`
	} `bson:"originTx"`
}

// FindSourceTxById returns the source transaction document with the given ID.
func (r *MongoRepository) FindSourceTxById(ctx context.Context, id string) (*SourceTxDoc, error) {
	var sourceTxDoc SourceTxDoc
	err := r.globalTransactions.FindOne(ctx, bson.M{"_id": id}).Decode(&sourceTxDoc)
	if err != nil {
		return nil, err
	}
	return &sourceTxDoc, err
}

// GetIDByVaaID returns the id for the given vaa id
func (p *MongoRepository) GetIDByVaaID(ctx context.Context, vaaID string) (string, error) {
	vaa, err := p.vaaRepository.FindById(ctx, vaaID)
	if err != nil {
		return "", err
	}
	return domain.GetDigestFromRaw(vaa.Vaa)
}
