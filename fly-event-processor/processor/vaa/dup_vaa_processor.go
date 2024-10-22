package vaa

import (
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type DuplicateVaaProcessor struct {
	guardianPool *pool.Pool
	repository   *storage.Repository
	logger       *zap.Logger
	metrics      metrics.Metrics
}

func NewDuplicateVaaProcessor(
	guardianPool *pool.Pool,
	repository *storage.Repository,
	logger *zap.Logger,
	metrics metrics.Metrics,
) *DuplicateVaaProcessor {
	return &DuplicateVaaProcessor{
		guardianPool: guardianPool,
		repository:   repository,
		logger:       logger,
		metrics:      metrics,
	}
}

func (p *DuplicateVaaProcessor) Process(ctx context.Context, params *Params) error {
	logger := p.logger.With(
		zap.String("trackId", params.TrackID),
		zap.String("vaaId", params.VaaID))

	// 1. check if the vaa stored in the VAA collections is the correct one.

	// 1.1 get vaa from Vaas collection
	vaaDoc, err := p.repository.FindVAAById(ctx, params.VaaID)
	if err != nil {
		logger.Error("error getting vaa from collection", zap.Error(err))
		return err
	}

	// 1.2 if the event time has not reached the finality time, the event fail and
	// will be reprocesed on the next retry.
	finalityTime := getFinalityTimeByChainID(params.ChainID)
	if vaaDoc.Timestamp == nil {
		logger.Error("vaa timestamp is nil")
		return errors.New("vaa timestamp is nil")
	}

	vaaTimestamp := *vaaDoc.Timestamp
	reachedFinalityTime := time.Now().After(vaaTimestamp.Add(finalityTime))
	if !reachedFinalityTime {
		logger.Debug("event time has not reached the finality time",
			zap.Time("finalityTime", vaaTimestamp.Add(finalityTime)))
		return errors.New("event time has not reached the finality time")
	}

	// 1.3 call guardian api to get signed_vaa.
	guardians := p.guardianPool.GetItems()
	var signedVaa *guardian.SignedVaa
	for _, g := range guardians {
		g.Wait(ctx)
		guardianAPIClient, err := guardian.NewGuardianAPIClient(
			guardian.DefaultTimeout,
			g.Id,
			logger)
		if err != nil {
			logger.Error("error creating guardian api client", zap.Error(err))
			continue
		}
		signedVaa, err = guardianAPIClient.GetSignedVAA(params.VaaID)
		if err != nil {
			logger.Error("error getting signed vaa from guardian api", zap.Error(err))
			continue
		}
		break
	}

	if signedVaa == nil {
		logger.Error("error getting signed vaa from guardian api")
		return errors.New("error getting signed vaa from guardian api")
	}

	// 1.4 compare digest from vaa and signedVaa
	guardianVAA, err := sdk.Unmarshal(signedVaa.VaaBytes)
	if err != nil {
		logger.Error("error unmarshalling guardian signed vaa", zap.Error(err))
		return err
	}

	vaa, err := sdk.Unmarshal(vaaDoc.Vaa)
	if err != nil {
		logger.Error("error unmarshalling vaa", zap.Error(err))
		return err
	}

	// If the guardian digest is the same that the vaa digest,
	// the stored vaa in the vaas collection is the correct one.
	if guardianVAA.HexDigest() == vaa.HexDigest() {
		logger.Info("vaa stored in vaas collections is the correct")
		return nil
	}

	// 2. Check for each duplicate VAAs to detect which is the correct one.

	// 2.1 This check is necessary to avoid race conditions when the vaa is processed
	if vaaDoc.TxHash == "" {
		logger.Error("vaa txHash is empty")
		return errors.New("vaa txHash is empty")
	}

	// 2.2 Get all duplicate vaas by vaaId
	duplicateVaaDocs, err := p.repository.FindDuplicateVAAs(ctx, params.VaaID)
	if err != nil {
		logger.Error("error getting duplicate vaas from collection", zap.Error(err))
		return err
	}

	// 2.3 Check each duplicate VAA to detect which is the correct one.
	for _, duplicateVaaDoc := range duplicateVaaDocs {
		duplicateVaa, err := sdk.Unmarshal(duplicateVaaDoc.Vaa)
		if err != nil {
			logger.Error("error unmarshalling vaa", zap.Error(err))
			return err
		}

		if guardianVAA.HexDigest() == duplicateVaa.HexDigest() {
			err := p.repository.FixVAA(ctx, params.VaaID, duplicateVaaDoc.ID)
			if err != nil {
				logger.Error("error fixing vaa", zap.Error(err))
				return err
			}
			logger.Info("vaa fixed")
			return nil
		}
	}

	logger.Info("can't fix duplicate vaa")
	p.metrics.IncDuplicatedVaaCanNotFixed(params.ChainID, config.DbLayerMongo)
	return errors.New("can't fix duplicate vaa")
}
