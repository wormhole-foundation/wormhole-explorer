package guardiansets

import (
	"context"
	"errors"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.uber.org/zap"
)

var ErrGuardianSetNotFound = errors.New("guardian set not found")

type mongoGuardianSet struct {
	ethGuardianSet    *ethGuardianSet
	repository        repository.GuardianSetStorager
	manualGuardianSet *manualGuardianSet
	logger            *zap.Logger
}

func NewMongoGuardianSet(ethGuardianSet *ethGuardianSet, repository repository.GuardianSetStorager,
	manualGuardianSet *manualGuardianSet, logger *zap.Logger) *mongoGuardianSet {
	return &mongoGuardianSet{
		ethGuardianSet:    ethGuardianSet,
		repository:        repository,
		manualGuardianSet: manualGuardianSet,
		logger:            logger,
	}
}

func (m *mongoGuardianSet) Sync(ctx context.Context) error {
	firstIndex := uint32(0)
	mongoGuardianSetIndex, err := m.GetCurrentGuardianSetIndex(ctx)
	if err != nil {
		if err != ErrGuardianSetNotFound {
			return err
		}
	} else {
		firstIndex = mongoGuardianSetIndex + 1
	}

	ethGuardianSetIndex, err := m.ethGuardianSet.GetCurrentGuardianSetIndex(ctx)
	if err != nil {
		return err
	}

	for index := firstIndex; index <= ethGuardianSetIndex; index++ {
		m.logger.Info("Syncing guardian set", zap.Uint32("index", index))
		// Get manual set from config
		guardianSet, expirationTime, err := m.manualGuardianSet.GetGuardianSet(ctx, index)
		if err != nil {
			return err
		}
		if guardianSet != nil {
			// Save to mongo
			err = m.Upsert(ctx, guardianSet, expirationTime)
			if err != nil {
				return err
			}
			continue
		}
		// Get guardian set from eth
		guardianSet, expirationTime, err = m.ethGuardianSet.GetGuardianSet(ctx, index)
		if err != nil {
			return err
		}
		if guardianSet != nil {
			// Save to mongo
			err = m.Upsert(ctx, guardianSet, expirationTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *mongoGuardianSet) GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error) {
	list, err := m.repository.FindAll(ctx)
	if err != nil {
		return 0, err
	}
	if len(list) == 0 {
		return 0, ErrGuardianSetNotFound
	}
	maxIndex := uint32(0)
	for _, v := range list {
		if v.GuardianSetIndex > maxIndex {
			maxIndex = v.GuardianSetIndex
		}
	}
	return maxIndex, nil
}

func (m *mongoGuardianSet) Upsert(ctx context.Context, gst *common.GuardianSet, expiration *time.Time) error {
	var keys []repository.GuardianSetKey
	for index, v := range gst.Keys {
		keys = append(keys, repository.GuardianSetKey{
			Index:   uint32(index),
			Address: v.Bytes(),
		})
	}
	doc := &repository.GuardianSet{
		GuardianSetIndex: gst.Index,
		Keys:             keys,
		ExpirationTime:   expiration,
		UpdatedAt:        time.Now(),
	}
	return m.repository.Upsert(ctx, doc)
}
