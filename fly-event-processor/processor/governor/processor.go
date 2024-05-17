package governor

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type Processor struct {
	repository *storage.Repository
	logger     *zap.Logger
	metrics    metrics.Metrics
}

func NewProcessor(repository *storage.Repository, logger *zap.Logger, metrics metrics.Metrics) *Processor {
	return &Processor{
		repository: repository,
		logger:     logger,
		metrics:    metrics,
	}
}

func (p *Processor) Process(ctx context.Context, params *Params) error {
	logger := p.logger.With(
		zap.String("trackId", params.TrackID),
	)

	// 1. Check if the event is valid.
	if params.NodeGovernorVaa == nil {
		logger.Info("event is nil")
		return errors.New("event cannot be nil")
	}

	// 2. Get new and current governorVaa by node.
	newNodeGovernorVaa := params.NodeGovernorVaa
	currenGovernorVaaId, err := p.getCurrentGovernorVaa(ctx, newNodeGovernorVaa.Node.NodeAddress)
	if err != nil {
		logger.Error("failed to get current governorVaa",
			zap.Error(err),
			zap.String("nodeAddress", newNodeGovernorVaa.Node.NodeAddress))
		return err
	}

	// 3. Get nodeGovernorVaa to insert and delete.
	nodeGovernorVaaToInsert := p.getNodeGovernorVaaToInsert(ctx, newNodeGovernorVaa, currenGovernorVaaId)
	nodeGovernorVaaIdToDelete := p.getNodeGovernorVaaToDelete(ctx, newNodeGovernorVaa, currenGovernorVaaId)

	// 4. Get governorVaa to insert and delete.
	governorVaaToInsert := p.getGovernorVaaToInsert(ctx, nodeGovernorVaaToInsert)
	governorVaaIdToDelete, err := p.getGovernorVaaToDelete(ctx, newNodeGovernorVaa.NodeAddress, nodeGovernorVaaIdToDelete)
	if err != nil {
		logger.Error("failed to get governorVaa to delete",
			zap.Error(err),
			zap.String("nodeAddress", newNodeGovernorVaa.Node.NodeAddress))
		return err
	}

	// update nodeGovernorVaa and governorVaa records.
	// if err := p.updateGovernorVaa(ctx, nodeGovernorVaaToInsert, nodeGovernorVaaIdToDelete, governorVaaToInsert, governorVaaIdToDelete); err != nil {
	// 	logger.Error("failed to update governorVaa",
	// 		zap.Error(err),
	// 		zap.String("nodeAddress", newNodeGovernorVaa.Node.NodeAddress))
	// 	return err
	// }

	return nil
}

// getGovernorVaaToInsert gets the governor vaas to insert.
func (p *Processor) getGovernorVaaToInsert(ctx context.Context,
	nodeGovernorVaaToInsert *domain.NodeGovernorVaa) *domain.NodeGovernorVaa {

	// get vaaIDs to insert in nodeGovernorVaa
	vaaIds := make([]string, 0, len(nodeGovernorVaaToInsert.GovernorVaas))
	for vaaID := range nodeGovernorVaaToInsert.GovernorVaas {
		vaaIds = append(vaaIds, vaaID)
	}

	// get governorVaa by vaaIDs.
	governorVaaDocs, err := p.repository.FindGovernorVaaByVaaIDs(ctx, vaaIds)
	if err != nil {
		logger.Error("failed to find governorVaa by vaaIDs",
			zap.Error(err),
			zap.Strings("vaaIDs", vaaIds))
		return nil
	}

	allGovernorVaaInserted := len(nodeGovernorVaaToInsert.GovernorVaas) == len(governorVaaDocs)
	if allGovernorVaaInserted {
		return &domain.NodeGovernorVaa{
			Node:         nodeGovernorVaaToInsert.Node,
			GovernorVaas: map[string]domain.GovernorVaa{},
		}
	}

	// get governorVaa to insert
	governorVaaToInsert := make(map[string]domain.GovernorVaa)
	for _, g := nodeGovernorVaaToInsert.GovernorVaas{
		if _, ok := governorVaaDocs[g]; !ok {
			governorVaaToInsert[g] = nodeGovernorVaaToInsert.GovernorVaas[g]
	}

	}
}

// getGovernorVaaToDelete gets the governor vaas to delete.
func (p *Processor) getGovernorVaaToDelete(ctx context.Context,
	nodeAddress string, nodeGovernorVaaToDelete map[string]domain.Node) (map[string]domain.Node, error) {

	// get vaaIDs to delete in nodeGovernorVaa
	vaaIds := make([]string, 0, len(nodeGovernorVaaToDelete))
	for vaaID := range nodeGovernorVaaToDelete {
		vaaIds = append(vaaIds, vaaID)
	}

	// get nodeGovernorVaa by vaaIDs.
	nodeGovernorVaaDocs, err := p.repository.FindNodeGovernorVaaByVaaIDs(ctx, vaaIds)
	if err != nil {
		logger.Error("failed to find governorVaa by vaaIDs",
			zap.Error(err),
			zap.Strings("vaaIDs", vaaIds))
		return nil, err
	}

	// convert nodeGovernorVaaDocs to map[string][]*storage.NodeGovernorVaaDoc
	mapNodeGovernorVaaDoc := make(map[string][]*storage.NodeGovernorVaaDoc)
	for _, governorVaaDoc := range nodeGovernorVaaDocs {
		if _, ok := mapNodeGovernorVaaDoc[governorVaaDoc.VaaID]; !ok {
			mapNodeGovernorVaaDoc[governorVaaDoc.VaaID] = make([]*storage.NodeGovernorVaaDoc, 0)
		}
		mapNodeGovernorVaaDoc[governorVaaDoc.VaaID] = append(mapNodeGovernorVaaDoc[governorVaaDoc.VaaID], governorVaaDoc)
	}

	// get governorVaa to delete
	governorVaaToDelete := make(map[string]domain.Node)
	for vaaID, nodeGovernorVaa := range mapNodeGovernorVaaDoc {
		deleteGovernorVaa := len(nodeGovernorVaa) == 1 && nodeAddress == nodeGovernorVaa[0].NodeAddress
		if deleteGovernorVaa {
			governorVaaToDelete[vaaID] = domain.Node{
				NodeName:    nodeGovernorVaa[0].NodeName,
				NodeAddress: nodeGovernorVaa[0].NodeAddress,
			}
		}
	}

	return governorVaaToDelete, nil
}

// getNodeGovernorVaaToInsert gets the node governor vaas to insert.
func (p *Processor) getNodeGovernorVaaToInsert(_ context.Context, newGovernorVaa *domain.NodeGovernorVaa, currentGovernorVaa map[string]domain.Node) *domain.NodeGovernorVaa {
	nodeGovernorVaaToInsert := &domain.NodeGovernorVaa{
		Node:         newGovernorVaa.Node,
		GovernorVaas: make(map[string]domain.GovernorVaa),
	}

	for vaaID, newGovernorVaa := range newGovernorVaa.GovernorVaas {
		if _, ok := currentGovernorVaa[vaaID]; !ok {
			nodeGovernorVaaToInsert.GovernorVaas[vaaID] = newGovernorVaa
		}
	}
	return nodeGovernorVaaToInsert
}

// getNodeGovernorVaaToDelete gets the node governor vaas to delete.
func (p *Processor) getNodeGovernorVaaToDelete(ctx context.Context, newGovernorVaa *domain.NodeGovernorVaa, currentGovernorVaa map[string]domain.Node) map[string]domain.Node {
	nodeGovernorVaaToDelete := make(map[string]domain.Node)
	for vaaID, currentGovernorVaa := range currentGovernorVaa {
		if _, ok := newGovernorVaa.GovernorVaas[vaaID]; !ok {
			nodeGovernorVaaToDelete[vaaID] = currentGovernorVaa
		}
	}
	return nodeGovernorVaaToDelete
}

// getCurrentGovernorVaa gets the current governor vaa by nodeAddress.
func (p *Processor) getCurrentGovernorVaa(ctx context.Context, nodeAddress string) (map[string]domain.Node, error) {
	nodeGovernorVaaDoc, err := p.repository.FindNodeGovernorVaaByNodeAddress(ctx, nodeAddress)
	if err != nil {
		logger.Error("failed to find nodeGovernorVaa by nodeAddress",
			zap.Error(err),
			zap.String("nodeAddress", nodeAddress))
		return nil, err
	}
	mapCurrentNodeGovernorVaa := getMapGovernorVaa(nodeGovernorVaaDoc)
	return mapCurrentNodeGovernorVaa, nil
}

// getMapGovernorVaa converts []*storage.NodeGovernorVaa to map[vaaID]domain.Node
func getMapGovernorVaa(nodeGovernorVaas []*storage.NodeGovernorVaaDoc) map[string]domain.Node {
	mapGovernorVaa := make(map[string]domain.Node)
	for _, nodeGovernorVaa := range nodeGovernorVaas {
		mapGovernorVaa[nodeGovernorVaa.VaaID] = domain.Node{
			NodeName:    nodeGovernorVaa.NodeName,
			NodeAddress: nodeGovernorVaa.NodeAddress,
		}
	}
	return mapGovernorVaa
}

// addGovernorVaa adds new governor vaa for the node.
func (p *Processor) addGovernorVaa(ctx context.Context, governorVaa *domain.NodeGovernorVaa) error {
	return nil
}

// removeGovernorVaa removes governor vaa that are no longer in the node.
func (p *Processor) removeGovernorVaa(ctx context.Context, governorVaa *domain.NodeGovernorVaa) error {
	return nil
}
