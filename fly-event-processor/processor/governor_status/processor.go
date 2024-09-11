package governor_status

import (
	"context"
	"errors"
	"fmt"

	txTracker "github.com/wormhole-foundation/wormhole-explorer/common/client/txtracker"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
)

// Processor is a governor processor.
type Processor struct {
	repository       storage.GovernorStatusRepository
	createTxHashFunc txTracker.CreateTxHashFunc
	logger           *zap.Logger
	metrics          metrics.Metrics
}

// NewProcessor creates a new governor processor.
func NewProcessor(
	repository storage.GovernorStatusRepository,
	createTxHashFunc txTracker.CreateTxHashFunc,
	logger *zap.Logger,
	metrics metrics.Metrics,
) *Processor {

	return &Processor{
		repository:       repository,
		createTxHashFunc: createTxHashFunc,
		logger:           logger,
		metrics:          metrics,
	}
}

// Process processes a governor event.
func (p *Processor) Process(
	ctx context.Context,
	params *Params) error {

	logger := p.logger.With(
		zap.String("trackId", params.TrackID),
	)

	// 1. Check if the event is valid.
	if params.NodeGovernorVaa == nil {
		logger.Info("event is nil")
		return errors.New("event cannot be nil")
	}
	node := params.NodeGovernorVaa.Node
	if node.Address == "" {
		logger.Info("node is invalid")
		return errors.New("node is invalid")
	}

	// 2. Get new and current governorVaa by node.
	newNodeGovernorVaas := params.NodeGovernorVaa
	nodeGovernorVaaIds, err := p.getNodeGovernorVaaIds(ctx, node, logger)
	if err != nil {
		logger.Error("failed to get current governorVaa",
			zap.Error(err),
			zap.String("nodeAddress", node.Address))
		return err
	}

	// 3. Get nodeGovernorVaa to add and delete.
	nodeGovernorVaasToAdd := getNodeGovernorVaasToAdd(
		newNodeGovernorVaas.GovernorVaas, nodeGovernorVaaIds)
	nodeGovernorVaaIdsToDelete := getNodeGovernorVaasToDelete(
		newNodeGovernorVaas.GovernorVaas, nodeGovernorVaaIds)

	// 4. Get governorVaa to add and delete.
	governorVaasToAdd, err := p.getGovernorVaaToAdd(ctx, nodeGovernorVaasToAdd, logger)
	if err != nil {
		logger.Error("failed to get governorVaa to insert",
			zap.Error(err),
			zap.String("nodeAddress", node.Address))
		return err
	}
	governorVaaIdsToDelete, err := p.getGovernorVaaToDelete(ctx, node, nodeGovernorVaaIdsToDelete, logger)
	if err != nil {
		logger.Error("failed to get governorVaa to delete",
			zap.Error(err),
			zap.String("nodeAddress", node.Address))
		return err
	}

	// 5. Check if there are no changes in governor.
	changeNodeGovernorVaas := len(nodeGovernorVaasToAdd) > 0 || len(nodeGovernorVaaIdsToDelete) > 0
	changeGovernorVaas := len(governorVaasToAdd) > 0 || len(governorVaaIdsToDelete) > 0
	if !changeNodeGovernorVaas && !changeGovernorVaas {
		logger.Info("no changes in governor",
			zap.String("nodeAddress", node.Address))
		return nil
	}

	// 6. Update governor data for the node.
	err = p.updateGovernorStatus(ctx,
		node,
		nodeGovernorVaasToAdd,
		nodeGovernorVaaIdsToDelete,
		governorVaasToAdd,
		governorVaaIdsToDelete)
	if err != nil {
		logger.Error("failed to update governorVaa",
			zap.Error(err),
			zap.String("nodeAddress", node.Address),
			zap.String("node", node.Name))
		return err
	}

	return nil
}

// getNodeGovernorVaaIds gets the current governor vaaIds stored in the database by node address.
func (p *Processor) getNodeGovernorVaaIds(
	ctx context.Context,
	node domain.Node,
	logger *zap.Logger,
) (Set[string], error) {

	// get current nodeGovernorVaa by nodeAddress.
	nodeGovernorVaaDoc, err := p.repository.FindNodeGovernorVaaByNodeAddress(ctx, node.Address)
	if err != nil {
		logger.Error("failed to find nodeGovernorVaa by nodeAddress",
			zap.Error(err),
			zap.String("nodeAddress", node.Address))
		return Set[string]{}, err
	}

	// convert nodeGovernorVaaDoc to Set[string]
	nodeGovernorVaaId := make(Set[string])
	for _, governorVaaDoc := range nodeGovernorVaaDoc {
		nodeGovernorVaaId.Add(governorVaaDoc.VaaID)
	}
	return nodeGovernorVaaId, nil
}

// getNodeGovernorVaasToAdd gets the node governor vaas to add.
func getNodeGovernorVaasToAdd(
	newNodeGovernorVaas map[string]domain.GovernorVaa,
	nodeGovernorVaaIds Set[string],
) map[string]domain.GovernorVaa {

	nodeGovernorVaasToAdd := make(map[string]domain.GovernorVaa)
	for vaaID, governorVaa := range newNodeGovernorVaas {
		if ok := nodeGovernorVaaIds.Contains(vaaID); !ok {
			nodeGovernorVaasToAdd[vaaID] = governorVaa
		}
	}
	return nodeGovernorVaasToAdd
}

// getNodeGovernorVaasToDelete gets the node governor vaas to delete.
func getNodeGovernorVaasToDelete(
	newNodeGovernorVaas map[string]domain.GovernorVaa,
	nodeGovernorVaaIds Set[string],
) Set[string] {

	nodeGovernorVaasToDelete := make(Set[string])
	for vaaID := range nodeGovernorVaaIds {
		if _, ok := newNodeGovernorVaas[vaaID]; !ok {
			nodeGovernorVaasToDelete.Add(vaaID)
		}
	}
	return nodeGovernorVaasToDelete
}

// getGovernorVaaToAdd gets the governor vaas to add.
func (p *Processor) getGovernorVaaToAdd(
	ctx context.Context,
	nodeGovernorVaas map[string]domain.GovernorVaa,
	logger *zap.Logger,
) ([]domain.GovernorVaa, error) {

	// get vaaIDs from the nodeGovernorVaas.
	vaaIds := make([]string, 0, len(nodeGovernorVaas))
	for vaaId, _ := range nodeGovernorVaas {
		vaaIds = append(vaaIds, vaaId)
	}

	// get governoVaas already added by vaaIDs.
	governorVaas, err := p.repository.FindGovernorVaaByVaaIDs(ctx, vaaIds)
	if err != nil {
		logger.Error("failed to find governor vaas by a list of vaaIDs",
			zap.Error(err),
			zap.Strings("vaaIDs", vaaIds))
		return nil, err
	}
	if len(vaaIds) < len(governorVaas) {
		logger.Error("failed to find governorVaa by a list of vaaIDs",
			zap.Error(err),
			zap.Strings("vaaIDs", vaaIds))
		return nil, errors.New("failed to find governorVaa by vaaIDs")
	}

	// check if all the governorVaa are already added
	if len(vaaIds) == len(governorVaas) {
		return nil, nil
	}

	// convert governorVaas to a set of vaaIDs.
	governorVaaIds := make(Set[string])
	for _, governorVaa := range governorVaas {
		governorVaaIds.Add(governorVaa.ID)
	}

	// get governorVaa to insert
	var governorVaasToInsert []domain.GovernorVaa
	for vaaID, governorVaa := range nodeGovernorVaas {
		if ok := governorVaaIds.Contains(vaaID); !ok {
			// fix governor vaa txHash
			txHash, err := p.createTxHashFunc(governorVaa.ID, governorVaa.TxHash)
			if err != nil {
				logger.Error("failed to create txHash",
					zap.Error(err),
					zap.String("vaaID", governorVaa.ID),
					zap.String("txHash", governorVaa.TxHash))
				return nil, err
			}
			governorVaa.TxHash = txHash.NativeTxHash
			governorVaasToInsert = append(governorVaasToInsert, governorVaa)
		}
	}

	return governorVaasToInsert, nil
}

// getGovernorVaaToDelete gets the governor vaas to delete.
func (p *Processor) getGovernorVaaToDelete(
	ctx context.Context,
	node domain.Node,
	nodeGovernorVaaIds Set[string],
	logger *zap.Logger,
) (Set[string], error) {

	// get vaaIDs from the nodeGovernorVaaIds.
	vaaIds := make([]string, 0, nodeGovernorVaaIds.Len())
	for vaaID := range nodeGovernorVaaIds {
		vaaIds = append(vaaIds, vaaID)
	}

	// nodeGovernorVaas contains all the node governor vaas that have the same vaaID.
	nodeGovernorVaas, err := p.repository.FindNodeGovernorVaaByVaaIDs(ctx, vaaIds)
	if err != nil {
		logger.Error("failed to find governorVaa by vaaIDs",
			zap.Error(err),
			zap.Strings("vaaIDs", vaaIds))
		return nil, err
	}

	// nodeAddressByVaaId contains all the node address grouped by vaaID.
	nodeAddressByVaaId := make(map[string][]string)
	for _, n := range nodeGovernorVaas {
		if _, ok := nodeAddressByVaaId[n.VaaID]; !ok {
			nodeAddressByVaaId[n.VaaID] = make([]string, 0)
		}
		nodeAddressByVaaId[n.VaaID] = append(nodeAddressByVaaId[n.VaaID], n.NodeAddress)
	}

	// get governorVaa to delete
	governorVaaToDelete := make(Set[string])
	for vaaID, nodeAddresses := range nodeAddressByVaaId {
		deleteGovernorVaa := len(nodeAddresses) == 1 && node.Address == nodeAddresses[0]
		if deleteGovernorVaa {
			governorVaaToDelete.Add(vaaID)
		}
	}

	return governorVaaToDelete, nil
}

func (p *Processor) updateGovernorStatus(ctx context.Context,
	node domain.Node,
	nodeGovernorVaasToAdd map[string]domain.GovernorVaa,
	nodeGovernorVaaIdsToDelete Set[string],
	governorVaasToAdd []domain.GovernorVaa,
	governorVaaIdsToDelete Set[string]) error {

	// convert nodeGovernorVaasToAdd to []storage.NodeGovernorVaaDoc
	var nodeGovernorVaasToAddDoc []storage.NodeGovernorVaa
	for vaaID, _ := range nodeGovernorVaasToAdd {
		nodeGovernorVaasToAddDoc = append(nodeGovernorVaasToAddDoc, storage.NodeGovernorVaa{
			ID:          fmt.Sprintf("%s-%s", node.Address, vaaID),
			NodeName:    node.Name,
			NodeAddress: node.Address,
			VaaID:       vaaID,
		})
	}

	// convert governorVaasToAdd to []storage.GovernorVaaDoc
	var governorVaasToAddDoc []storage.GovernorVaa
	for _, governorVaa := range governorVaasToAdd {
		governorVaasToAddDoc = append(governorVaasToAddDoc, storage.GovernorVaa{
			ID:             governorVaa.ID,
			ChainID:        governorVaa.ChainID,
			EmitterAddress: governorVaa.EmitterAddress,
			Sequence:       governorVaa.Sequence,
			TxHash:         governorVaa.TxHash,
			ReleaseTime:    governorVaa.ReleaseTime,
			Amount:         storage.Uint64(governorVaa.Amount),
		})
	}

	// convert nodeGovernorVaas vaaIds to ids
	var nodeGovVaaIdsToDelete []string
	for vaaID := range nodeGovernorVaaIdsToDelete {
		nodeGovVaaIdsToDelete = append(nodeGovVaaIdsToDelete, fmt.Sprintf("%s-%s", node.Address, vaaID))
	}

	return p.repository.UpdateGovernorStatus(
		ctx,
		node.Name,
		node.Address,
		nodeGovernorVaasToAddDoc,
		nodeGovVaaIdsToDelete,
		governorVaasToAddDoc,
		governorVaaIdsToDelete.ToSlice())
}
