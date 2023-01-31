// Package observations handle the request of VAA data from governor endpoint defined in the api.
package vaa

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *vaa.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(serv *vaa.Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "VaaController"))}
}

// FindSignedVAAByID godoc
// @Description get a VAA []byte from a chainID, emitter address and sequence.
// @Tags Guardian
// @ID guardians-find-signed-vaa
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} object{vaaBytes=[]byte}
// @Failure 400
// @Failure 500
// @Router /v1/signed_vaa/{chain_id}/{emitter}/{seq} [get]
func (c *Controller) FindSignedVAAByID(ctx *fiber.Ctx) error {
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}

	// TODO
	// check chainID is not Pyth. Pyth message are not stored with the other vaa.
	//if ChainIDPythNet == chainID {
	//	return response.NewApiError(ctx, fiber.StatusBadRequest, response.InvalidParam,
	//		"not supported for PythNet", nil)
	//}
	vaa, err := c.srv.FindById(ctx.Context(), chainID, *emitter, strconv.FormatUint(seq, 10), false, "")
	if err != nil {
		return err
	}
	response := struct {
		VaaBytes []byte `json:"vaaBytes"`
	}{
		VaaBytes: vaa.Data.Vaa,
	}
	return ctx.JSON(response)
}

// FindSignedBatchVAAByID godoc
// @Description get a batch of VAA []byte from a chainID, emitter address and sequence.
// @Tags Guardian
// @ID guardians-find-signed-batch-vaa
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} object{vaaBytes=[]byte}
// @Failure 400
// @Failure 500
// @Router /v1/signed_batch_vaa/{chain_id}/{emitter}/sequence/{seq} [get]
func (c *Controller) FindSignedBatchVAAByID(ctx *fiber.Ctx) error {
	return response.NewApiError(ctx, fiber.StatusNotImplemented, response.Unimplemented, "not yet implemented", nil)
}
