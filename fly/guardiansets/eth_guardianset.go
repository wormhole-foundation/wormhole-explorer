package guardiansets

import (
	"context"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/watchers/evm/connectors"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"go.uber.org/zap"
)

type ethGuardianSet struct {
	contractAddress ethCommon.Address
	alertClient     alert.AlertClient
	connector       *connectors.EthereumBaseConnector
	logger          *zap.Logger
}

var _ GuardianSetProvider = &ethGuardianSet{}

func NewEthGuardianSet(ctx context.Context, contract, url string, alertClient alert.AlertClient, logger *zap.Logger) (*ethGuardianSet, error) {
	contractAddress := ethCommon.HexToAddress(contract)
	connector, err := connectors.NewEthereumBaseConnector(ctx, "ethereum", url, contractAddress, logger)
	if err != nil {
		return nil, err
	}
	return &ethGuardianSet{
		contractAddress: contractAddress,
		alertClient:     alertClient,
		connector:       connector,
		logger:          logger,
	}, nil
	//return nil, nil
}

func (e *ethGuardianSet) GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error) {
	e.logger.Debug("Fetching current eth guardian set index")
	return e.connector.GetCurrentGuardianSetIndex(ctx)
	//return 0, nil
}

func (e *ethGuardianSet) GetGuardianSet(ctx context.Context, index uint32) (*common.GuardianSet, *time.Time, error) {
	e.logger.Debug("Fetching eth guardian set", zap.Uint32("index", index))
	guardianSet, err := e.connector.GetGuardianSet(ctx, index)
	if err != nil {
		return nil, nil, err
	}
	var expirationTime *time.Time
	if guardianSet.ExpirationTime != 0 {
		et := time.Unix(int64(guardianSet.ExpirationTime), 0)
		expirationTime = &et
	}
	return &common.GuardianSet{Index: index, Keys: guardianSet.Keys}, expirationTime, nil
	//return nil, nil, nil
}
func (e *ethGuardianSet) GetGuardianSetHistory(ctx context.Context) (*GuardianSetHistory, error) {
	guardianSetIndex, err := e.connector.GetCurrentGuardianSetIndex(ctx)
	if err != nil {
		return nil, err
	}

	var guardianSetsByIndex []common.GuardianSet
	var expirationTimesByIndex []time.Time
	for index := uint32(0); index <= guardianSetIndex; index++ {
		guardianSet, expirationTime, err := e.GetGuardianSet(ctx, index)
		if err != nil {
			return nil, err
		}
		guardianSetsByIndex = append(guardianSetsByIndex, *guardianSet)
		var et time.Time
		if expirationTime != nil {
			et = *expirationTime
		}
		expirationTimesByIndex = append(expirationTimesByIndex, et)
	}
	return &GuardianSetHistory{
		guardianSetsByIndex:    guardianSetsByIndex,
		expirationTimesByIndex: expirationTimesByIndex,
		alertClient:            e.alertClient,
	}, nil
	//return nil, nil
}

// AddGuardianSet implements GuardianSetProvider.
func (e *ethGuardianSet) AddGuardianSet(ctx context.Context, gs *common.GuardianSet, et time.Time) error {
	return nil
}
