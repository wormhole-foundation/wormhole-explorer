package guardiansets

import (
	"context"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"go.uber.org/zap"
)

type GuardianSetSynchronizer struct {
	provider   GuardianSetProvider
	gst        *common.GuardianSetState
	gstHistory *GuardianSetHistory
	logger     *zap.Logger
}

// NewGuardianSetSynchronizer creates a new GuardianSetSynchronizer.
func NewGuardianSetSynchronizer(ctx context.Context, gst *common.GuardianSetState, provider GuardianSetProvider, logger *zap.Logger) (*GuardianSetSynchronizer, error) {
	guardianSetHistory, err := provider.GetGuardianSetHistory(ctx)
	if err != nil {
		return nil, err
	}
	gsLastet := guardianSetHistory.GetLatest()
	gst.Set(&gsLastet)
	return &GuardianSetSynchronizer{
		provider:   provider,
		gst:        gst,
		gstHistory: guardianSetHistory,
		logger:     logger,
	}, nil
}

// Sync synchronizes the guardian set state with the provider.
func (s *GuardianSetSynchronizer) Sync(ctx context.Context) {
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				currentIndex := s.gst.Get().Index
				index, err := s.provider.GetCurrentGuardianSetIndex(ctx)
				if err != nil {
					s.logger.Error("failed to get current guardian set index", zap.Error(err))
					return
				}
				s.logger.Info("current guardian set index", zap.Uint32("index", index))
				if index > currentIndex {
					gs, expiration, err := s.provider.GetGuardianSet(ctx, index)
					if err != nil {
						s.logger.Error("failed to get guardian set", zap.Error(err))
						return
					}
					var et time.Time
					if expiration != nil {
						et = *expiration
					}
					s.gst.Set(gs)
					s.gstHistory.Add(*gs, et)
					s.provider.AddGuardianSet(ctx, gs, et)
					s.logger.Info("guardian set updated", zap.Uint32("index", index))
				}
			}
		}
	}()
}

func (s *GuardianSetSynchronizer) GetLatestGuardianSet() *common.GuardianSetState {
	return s.gst

}

func (s *GuardianSetSynchronizer) GetGuardianSetHistory() *GuardianSetHistory {
	return s.gstHistory
}
