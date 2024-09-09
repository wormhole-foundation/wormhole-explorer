package storage

import (
	"context"

	commonRepo "github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

// Repository exposes operations over the `globalTransactions` collection.
// Remove Repository after switch to use only postgres.
type Repository struct {
	logger           *zap.Logger
	metrics          metrics.Metrics
	vaas             *mongo.Collection
	duplicateVaas    *mongo.Collection
	nodeGovernorVaas *mongo.Collection
	governorVaas     *mongo.Collection
}

// New creates a new repository.
func NewRepository(logger *zap.Logger, db *mongo.Database, metrics metrics.Metrics) *Repository {
	r := Repository{
		logger:           logger,
		metrics:          metrics,
		vaas:             db.Collection(commonRepo.Vaas),
		duplicateVaas:    db.Collection(commonRepo.DuplicateVaas),
		nodeGovernorVaas: db.Collection(commonRepo.NodeGovernorVaas),
		governorVaas:     db.Collection(commonRepo.GovernorVaas),
	}
	return &r
}

// FindVAAById find a vaa by id.
func (r *Repository) FindVAAById(ctx context.Context, vaaID string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": vaaID}).Decode(&vaaDoc)
	return &vaaDoc, err
}

// FindDuplicateVAAById find a duplicate vaa by id.
func (r *Repository) FindDuplicateVAAById(ctx context.Context, id string) (*DuplicateVaaDoc, error) {
	var duplicateVaaDoc DuplicateVaaDoc
	err := r.duplicateVaas.FindOne(ctx, bson.M{"_id": id}).Decode(&duplicateVaaDoc)
	return &duplicateVaaDoc, err
}

// FindDuplicateVAAs find duplicate vaas by vaa id.
func (r *Repository) FindDuplicateVAAs(ctx context.Context, vaaID string) ([]DuplicateVaaDoc, error) {
	var duplicateVaaDocs []DuplicateVaaDoc
	cursor, err := r.duplicateVaas.Find(ctx, bson.M{"vaaId": vaaID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &duplicateVaaDocs); err != nil {
		return nil, err
	}
	return duplicateVaaDocs, nil
}

// FixVAA fix a vaa by id.
func (r *Repository) FixVAA(ctx context.Context, vaaID, duplicateID string) error {
	// start mongo transaction
	session, err := r.vaas.Database().Client().StartSession()
	if err != nil {
		return err
	}

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// get VAA by id
	vaaDoc, err := r.FindVAAById(ctx, vaaID)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// get duplicate vaa by id
	duplicateVaaDoc, err := r.FindDuplicateVAAById(ctx, duplicateID)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// create new vaa and new duplicate vaa
	newVaa := duplicateVaaDoc.ToVaaDoc(true)
	newDuplicateVaa, err := vaaDoc.ToDuplicateVaaDoc()
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// remove vaa
	_, err = r.vaas.DeleteOne(ctx, bson.M{"_id": vaaID})
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// remove duplicate vaa
	_, err = r.duplicateVaas.DeleteOne(ctx, bson.M{"_id": duplicateID})
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// insert new vaa
	_, err = r.vaas.InsertOne(ctx, newVaa)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// insert new duplicate vaa
	_, err = r.duplicateVaas.InsertOne(ctx, newDuplicateVaa)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	// commit transaction
	err = session.CommitTransaction(ctx)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	return nil
}

// FindNodeGovernorVaaByNodeAddress find governor vaas by node address.
func (r *Repository) FindNodeGovernorVaaByNodeAddress(ctx context.Context, nodeAddress string) ([]NodeGovernorVaa, error) {
	var nodeGovernorVaa []NodeGovernorVaa
	cursor, err := r.nodeGovernorVaas.Find(ctx, bson.M{"nodeAddress": nodeAddress})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &nodeGovernorVaa); err != nil {
		return nil, err
	}
	return nodeGovernorVaa, nil
}

// FindNodeGovernorVaaByVaaID find governor vaas by vaa id.
func (r *Repository) FindNodeGovernorVaaByVaaID(ctx context.Context, vaaID string) ([]NodeGovernorVaa, error) {
	var nodeGovernorVaa []NodeGovernorVaa
	cursor, err := r.nodeGovernorVaas.Find(ctx, bson.M{"vaaId": vaaID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &nodeGovernorVaa); err != nil {
		return nil, err
	}
	return nodeGovernorVaa, nil
}

// FindNodeGovernorVaaByVaaIDs find governor vaas by vaa ids.
func (r *Repository) FindNodeGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]NodeGovernorVaa, error) {
	var nodeGovernorVaa []NodeGovernorVaa
	cursor, err := r.nodeGovernorVaas.Find(ctx, bson.M{"vaaId": bson.M{"$in": vaaID}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &nodeGovernorVaa); err != nil {
		return nil, err
	}
	return nodeGovernorVaa, nil
}

// FindGovernorVaaByVaaID find governor vaas by a list of vaaIds
func (r *Repository) FindGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]GovernorVaa, error) {
	var governorVaa []GovernorVaa
	cursor, err := r.governorVaas.Find(ctx, bson.M{"_id": bson.M{"$in": vaaID}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &governorVaa); err != nil {
		return nil, err
	}
	return governorVaa, nil
}

func (r *Repository) UpdateGovernorStatus(
	ctx context.Context,
	nodeName string,
	nodeAddress string,
	nodeGovernorVaaDocToInsert []NodeGovernorVaa,
	nodeGovernorVaaDocToDelete []string,
	governorVaasToInsert []GovernorVaa,
	governorVaaIdsToDelete []string) error {
	err := r.updateGovernorStatus(ctx,
		nodeGovernorVaaDocToInsert,
		nodeGovernorVaaDocToDelete,
		governorVaasToInsert,
		governorVaaIdsToDelete)
	if err != nil {
		r.metrics.IncGovernorStatusUpdateFailed(
			nodeName, nodeAddress, config.DbLayerMongo)
	}
	return err
}

func (r *Repository) updateGovernorStatus(
	ctx context.Context,
	nodeGovernorVaaDocToInsert []NodeGovernorVaa,
	nodeGovernorVaaDocToDelete []string,
	governorVaasToInsert []GovernorVaa,
	governorVaaIdsToDelete []string) error {

	// 1. start mongo transaction
	session, err := r.vaas.Database().Client().StartSession()
	if err != nil {
		return err
	}

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// 2. insert node governor vaas.
	if len(nodeGovernorVaaDocToInsert) > 0 {
		var nodeGovVaadocs []interface{}
		for _, doc := range nodeGovernorVaaDocToInsert {
			nodeGovVaadocs = append(nodeGovVaadocs, doc)
		}
		_, err = r.nodeGovernorVaas.InsertMany(ctx, nodeGovVaadocs)
		if err != nil {
			session.AbortTransaction(ctx)
			return err
		}
	}

	// 3. delete node governor vaas.
	if len(nodeGovernorVaaDocToDelete) > 0 {
		_, err = r.nodeGovernorVaas.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": nodeGovernorVaaDocToDelete}})
		if err != nil {
			session.AbortTransaction(ctx)
			return err
		}
	}

	// 4. insert governor vaas.
	if len(governorVaasToInsert) > 0 {
		var govVaaDocs []interface{}
		for _, doc := range governorVaasToInsert {
			govVaaDocs = append(govVaaDocs, doc)
		}
		_, err = r.governorVaas.InsertMany(ctx, govVaaDocs)
		if err != nil {
			session.AbortTransaction(ctx)
			return err
		}
	}

	// 5. delete governor vaas.
	if len(governorVaaIdsToDelete) > 0 {
		_, err = r.governorVaas.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": governorVaaIdsToDelete}})
		if err != nil {
			session.AbortTransaction(ctx)
			return err
		}
	}

	// 6. commit transaction
	err = session.CommitTransaction(ctx)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	return nil
}
