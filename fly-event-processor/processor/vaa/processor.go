package vaa

import (
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Processor is a processor.
type Processor struct {
	guardianPool *pool.Pool
	repository   *storage.PostgresRepository
	logger       *zap.Logger
	metrics      metrics.Metrics
}

// NewProcessor creates a new processor.
func NewProcessor(
	guardianPool *pool.Pool,
	repository *storage.PostgresRepository,
	logger *zap.Logger,
	metrics metrics.Metrics,
) *Processor {
	return &Processor{
		guardianPool: guardianPool,
		repository:   repository,
		logger:       logger,
		metrics:      metrics,
	}
}

// Process processes a vaa message.
func (p *Processor) Process(ctx context.Context, params *Params) error {
	logger := p.logger.With(
		zap.String("trackId", params.TrackID),
		zap.String("vaaId", params.VaaID))

	// 1. check if the vaa stored in the VAA collections as actve is the correct one.

	// 1.1 get active attestation vaas from wh_attestation_vaas table.
	attestationVaa, err := p.repository.FindActiveAttestationVaaByVaaID(ctx, params.VaaID)
	if err != nil {
		logger.Error("error getting attestation vaas", zap.Error(err))
		return err
	}

	// 1.2 if the event time has not reached the finality time, the event fail and
	// will be reprocesed on the next retry.
	finalityTime := getFinalityTimeByChainID(params.ChainID)

	reachedFinalityTime := time.Now().After(attestationVaa.Timestamp.Add(finalityTime))
	if !reachedFinalityTime {
		logger.Debug("event time has not reached the finality time",
			zap.Time("finalityTime", attestationVaa.Timestamp.Add(finalityTime)))
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

	vaa, err := sdk.Unmarshal(attestationVaa.Raw)
	if err != nil {
		logger.Error("error unmarshalling vaa", zap.Error(err))
		return err
	}

	// If the guardian digest is the same that the vaa digest,
	// the stored vaa in the vaas collection is the correct one.
	if guardianVAA.HexDigest() == vaa.HexDigest() {
		if !attestationVaa.IsDuplicated {
			p.repository.UpdateAttestationVaaIsDuplicated(ctx, attestationVaa.ID, true)
		}
		logger.Info("vaa stored in vaas collections is the correct")
		return nil
	}

	// 2. Check for each duplicate VAAs to detect which is the correct one.

	attestationVaas, err := p.repository.FindAttestationVaaByVaaId(ctx, params.VaaID)
	if err != nil {
		logger.Error("error getting attestation vaas", zap.Error(err))
		return err
	}

	// 2.3 Check each duplicate VAA to detect which is the correct one.
	for _, v := range attestationVaas {
		vaa, err := sdk.Unmarshal(v.Raw)
		if err != nil {
			logger.Error("error unmarshalling vaa", zap.Error(err))
			return err
		}

		if guardianVAA.HexDigest() == vaa.HexDigest() {
			err := p.repository.FixActiveVaa(ctx, v.ID, params.VaaID)
			if err != nil {
				logger.Error("error fixing active vaa", zap.Error(err))
				return err
			}
			logger.Info("active vaa fixed",
				zap.String("vaaId", params.VaaID))
			return nil
		}
	}

	logger.Info("can't fix duplicate vaa")
	p.metrics.IncDuplicatedVaaCanNotFixed(params.ChainID)
	return errors.New("can't fix duplicate vaa")
}
