package vaa

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	logger           *zap.Logger
	rpcPool          map[sdk.ChainID]*pool.Pool
	wormchainRpcPool map[sdk.ChainID]*pool.Pool
	vaaRepository    vaa.VAARepository
	repository       consumer.Repository
	metrics          metrics.Metrics
	p2pNetwork       string
	notionalCache    *notional.NotionalCache
}

// NewController creates a Controller instance.
func NewController(rpcPool map[sdk.ChainID]*pool.Pool, wormchainRpcPool map[sdk.ChainID]*pool.Pool,
	vaaRepository vaa.VAARepository, repository consumer.Repository, p2pNetwork string, logger *zap.Logger,
	notionalCache *notional.NotionalCache) *Controller {
	return &Controller{
		metrics:          metrics.NewDummyMetrics(),
		rpcPool:          rpcPool,
		wormchainRpcPool: wormchainRpcPool,
		vaaRepository:    vaaRepository,
		repository:       repository,
		p2pNetwork:       p2pNetwork,
		logger:           logger,
		notionalCache:    notionalCache,
	}
}

func (c *Controller) Process(ctx *fiber.Ctx) error {
	var payload ProcessVaaRequest

	if err := ctx.BodyParser(&payload); err != nil {
		return err
	}

	c.logger.Info("Processing VAA from endpoint", zap.String("id", payload.ID))

	v, err := c.vaaRepository.GetVaa(ctx.Context(), payload.ID)
	if err != nil {
		return err
	}

	vaa, err := sdk.Unmarshal(v.Vaa)
	if err != nil {
		return err
	}

	digest, err := domain.GetDigestFromRaw(v.Vaa)
	if err != nil {
		return err
	}

	txHash := v.TxHash
	if txHash == "" {
		txHash, err = c.vaaRepository.GetTxHash(ctx.Context(), digest)
		if err != nil {
			return err
		}
	}

	p := &consumer.ProcessSourceTxParams{
		ID:          digest,
		TrackID:     fmt.Sprintf("controller-%s", payload.ID),
		Source:      "controller",
		Timestamp:   &vaa.Timestamp,
		VaaId:       vaa.MessageID(),
		ChainId:     vaa.EmitterChain,
		Emitter:     vaa.EmitterAddress.String(),
		Sequence:    strconv.FormatUint(vaa.Sequence, 10),
		TxHash:      txHash,
		Vaa:         v.Vaa,
		IsVaaSigned: true,
		Metrics:     c.metrics,
		Overwrite:   true,
		P2pNetwork:  c.p2pNetwork,
	}

	result, err := consumer.ProcessSourceTx(ctx.Context(), c.logger, c.rpcPool, c.wormchainRpcPool, c.repository, p, c.p2pNetwork, c.notionalCache)
	if err != nil {
		return err
	}

	return ctx.JSON(struct {
		Result any `json:"result"`
	}{Result: result})
}

func (c *Controller) CreateTxHash(ctx *fiber.Ctx) error {

	var payload TxHashRequest

	if err := ctx.BodyParser(&payload); err != nil {
		return err
	}

	txHash, err := hex.DecodeString(utils.Remove0x(payload.TxHash))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tx hash", "details": err.Error()})
	}

	c.logger.Info("Processing txHash from endpoint", zap.String("id", payload.VaaID))

	vaaID := strings.Split(payload.VaaID, "/")
	if len(vaaID) != 3 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid vaa id"})
	}

	chainIDStr, emitter, sequenceStr := vaaID[0], vaaID[1], vaaID[2]
	chainIDUint, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "chain id is not a number", "details": err.Error()})
	}
	chainID := sdk.ChainID(chainIDUint)
	if !domain.ChainIdIsValid(chainID) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chain id"})
	}

	encodedTxHash, err := domain.EncodeTrxHashByChainID(chainID, txHash)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tx hash", "details": err.Error()})
	}

	if chainID != sdk.ChainIDSolana && chainID != sdk.ChainIDAptos && chainID != sdk.ChainIDWormchain {
		return ctx.JSON(TxHashResponse{NativeTxHash: encodedTxHash})
	}

	sourceTx, err := c.repository.FindSourceTxById(ctx.Context(), payload.VaaID)
	if err == nil && sourceTx != nil {
		if sourceTx.OriginTx != nil && sourceTx.OriginTx.NativeTxHash != "" {
			return ctx.JSON(TxHashResponse{NativeTxHash: sourceTx.OriginTx.NativeTxHash})
		}
	}

	p := &consumer.ProcessSourceTxParams{
		TrackID:         "controller-tx-hash",
		Source:          "controller",
		Timestamp:       nil,
		VaaId:           payload.VaaID,
		ChainId:         chainID,
		Emitter:         emitter,
		Sequence:        sequenceStr,
		TxHash:          encodedTxHash,
		IsVaaSigned:     false,
		Metrics:         c.metrics,
		Overwrite:       true,
		DisableDBUpsert: true,
	}

	result, err := consumer.ProcessSourceTx(ctx.Context(), c.logger, c.rpcPool, c.wormchainRpcPool, c.repository, p, c.p2pNetwork, c.notionalCache)
	if err != nil {
		return err
	}

	return ctx.JSON(TxHashResponse{NativeTxHash: result.NativeTxHash})
}
