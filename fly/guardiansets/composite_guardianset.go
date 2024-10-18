package guardiansets

import (
	"context"
	"fmt"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

type compositeGuardianSet struct {
	ethGuardianSet    *ethGuardianSet
	dbGuardianSet     *dbGuardianSet
	manualGuardianSet *manualGuardianSet
	alertClient       alert.AlertClient
}

var _ GuardianSetProvider = &compositeGuardianSet{}

func NewCompositeGuardianSet(ethGuardianSet *ethGuardianSet, dbGuardianSet *dbGuardianSet,
	manualGuardianSet *manualGuardianSet, alertClient alert.AlertClient) *compositeGuardianSet {
	return &compositeGuardianSet{
		ethGuardianSet:    ethGuardianSet,
		dbGuardianSet:     dbGuardianSet,
		manualGuardianSet: manualGuardianSet,
		alertClient:       alertClient,
	}
}

func (c *compositeGuardianSet) Load(ctx context.Context) error {
	firstIndex := uint32(0)
	dbGuardianSetIndex, err := c.dbGuardianSet.GetCurrentGuardianSetIndex(ctx)
	if err != nil {
		if err != ErrGuardianSetNotFound {
			return err
		}
	} else {
		firstIndex = dbGuardianSetIndex + 1
	}

	ethGuardianSetIndex, err := c.ethGuardianSet.GetCurrentGuardianSetIndex(ctx)
	if err != nil {
		return err
	}

	for index := firstIndex; index <= ethGuardianSetIndex; index++ {

		guardianSet, expirationTime, err := c.manualGuardianSet.GetGuardianSet(ctx, index)
		if err != nil {
			return err
		}
		if guardianSet != nil {
			// Save to mongo
			err = c.dbGuardianSet.Upsert(ctx, guardianSet, expirationTime)
			if err != nil {
				return err
			}
			continue
		}

		guardianSet, expirationTime, err = c.ethGuardianSet.GetGuardianSet(ctx, index)
		if err != nil {
			return err
		}
		if guardianSet != nil {
			// Save to mongo
			err = c.dbGuardianSet.Upsert(ctx, guardianSet, expirationTime)
			if err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("guardian set not found for index %d", index)
	}
	return nil
}

func (e *compositeGuardianSet) GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error) {
	return e.ethGuardianSet.GetCurrentGuardianSetIndex(ctx)
}

func (e *compositeGuardianSet) GetGuardianSet(ctx context.Context, index uint32) (*common.GuardianSet, *time.Time, error) {
	guardianSet, expirationTime, _ := e.manualGuardianSet.GetGuardianSet(ctx, index)
	if guardianSet != nil {
		return guardianSet, expirationTime, nil
	}
	guardianSet, expirationTime, _ = e.dbGuardianSet.ethGuardianSet.GetGuardianSet(ctx, index)
	if guardianSet != nil {
		return guardianSet, expirationTime, nil
	}
	return e.ethGuardianSet.GetGuardianSet(ctx, index)
}

func (e *compositeGuardianSet) GetGuardianSetHistory(ctx context.Context) (*GuardianSetHistory, error) {
	guardianSetIndex, err := e.GetCurrentGuardianSetIndex(ctx)
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

}

func (e *compositeGuardianSet) AddGuardianSet(ctx context.Context, gs *common.GuardianSet, et time.Time) error {
	return e.dbGuardianSet.Upsert(ctx, gs, &et)
}
