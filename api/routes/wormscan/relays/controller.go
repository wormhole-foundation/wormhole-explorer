package relays

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/relays"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *relays.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(srv *relays.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "RelaysController")),
	}
}

// FindByVAA godoc
// @Description Get a specific relay information by chainID, emitter address and sequence.
// @Tags wormholescan
// @ID find-relay-by-vaa-id
// @Success 200 {object} relays.RelayResponse
// @Failure 400
// @Failure 500
// @Router /api/v1/relays/:chain/:emitter/:sequence [get]
func (c *Controller) FindOne(ctx *fiber.Ctx) error {
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}
	relay, err := c.srv.FindByVAA(ctx.Context(), middleware.UsePostgres(ctx), chainID, addr, strconv.FormatUint(seq, 10))
	if err != nil {
		return err
	}
	response := c.makeResponse(relay)
	return ctx.JSON(response)
}

func (c *Controller) makeResponse(doc *relays.RelayDoc) *RelayResponse {
	var data *RelayDataResponse
	if doc.Data.Metadata != nil {
		data = &RelayDataResponse{
			FromTxHash:  doc.Data.FromTxHash,
			ToTxHash:    doc.Data.ToTxHash,
			MaxAttempts: doc.Data.Metadata.MaxAttempts,
			Delivery: DeliveryReponse{
				ResultExecution: ResultExecutionResponse{
					TransactionHash: doc.Data.Metadata.DeliveryRecord.ResultLog.TransactionHash,
					RefundStatus:    doc.Data.Metadata.DeliveryRecord.ResultLog.RefundStatus,
					RevertString:    doc.Data.Metadata.DeliveryRecord.ResultLog.RevertString,
					Status:          doc.Data.Metadata.DeliveryRecord.ResultLog.Status,
					GasUsed:         doc.Data.Metadata.DeliveryRecord.ResultLog.GasUsed,
					Detail:          doc.Data.Metadata.DeliveryRecord.ResultString,
				},
				RelayGasUsed:        doc.Data.Metadata.DeliveryRecord.GasUsed,
				MaxRefund:           doc.Data.Metadata.DeliveryRecord.MaxRefund,
				Budget:              doc.Data.Metadata.DeliveryRecord.Budget,
				TargetChainDecimals: doc.Data.Metadata.DeliveryRecord.TargetChainDecimals,
			},
			Instructions: InstructionsResponse{
				EncodedExecutionInfo:   doc.Data.Metadata.Instructions.EncodedExecutionInfo,
				RefundAddress:          doc.Data.Metadata.Instructions.RefundAddress,
				SourceDeliveryProvider: doc.Data.Metadata.Instructions.SourceDeliveryProvider,
				SenderAddress:          doc.Data.Metadata.Instructions.SenderAddress,
				VaaKeys:                doc.Data.Metadata.Instructions.VaaKeys,
				ExtraReceiverValue: struct {
					Hex         string `json:"_hex"`
					IsBigNumber bool   `json:"_isBigNumber"`
				}{
					Hex:         doc.Data.Metadata.Instructions.ExtraReceiverValue.Hex,
					IsBigNumber: doc.Data.Metadata.Instructions.ExtraReceiverValue.IsBigNumber,
				},
				TargetAddress: doc.Data.Metadata.Instructions.TargetAddress,
				RequestedReceiverValue: struct {
					Hex         string `json:"_hex"`
					IsBigNumber bool   `json:"_isBigNumber"`
				}{
					Hex:         doc.Data.Metadata.Instructions.RequestedReceiverValue.Hex,
					IsBigNumber: doc.Data.Metadata.Instructions.RequestedReceiverValue.IsBigNumber,
				},
				RefundChainID:          doc.Data.Metadata.Instructions.RefundChainID,
				RefundDeliveryProvider: doc.Data.Metadata.Instructions.RefundDeliveryProvider,
				TargetChainID:          doc.Data.Metadata.Instructions.TargetChainID,
			},
		}
	}
	return &RelayResponse{
		ID:          doc.ID,
		Relayer:     doc.Origin,
		ReceivedAt:  doc.Data.ReceivedAt,
		Status:      doc.Data.Status,
		CompletedAt: doc.Data.CompletedAt,
		FailedAt:    doc.Data.FailedAt,
		Data:        data,
	}
}

type RelayResponse struct {
	ID          string             `json:"id"`
	Relayer     string             `json:"relayer"`
	Status      string             `json:"status"`
	ReceivedAt  time.Time          `json:"receivedAt"`
	CompletedAt *time.Time         `json:"completedAt"`
	FailedAt    *time.Time         `json:"failedAt"`
	Data        *RelayDataResponse `json:"data"`
}

type RelayDataResponse struct {
	FromTxHash   string               `json:"fromTxHash"`
	ToTxHash     *string              `json:"toTxHash"`
	MaxAttempts  int                  `json:"maxAttempts"`
	Instructions InstructionsResponse `json:"instructions"`
	Delivery     DeliveryReponse      `json:"delivery"`
}

type DeliveryReponse struct {
	ResultExecution     ResultExecutionResponse `json:"execution"`
	RelayGasUsed        int                     `json:"relayGasUsed"`
	MaxRefund           string                  `json:"maxRefund"`
	Budget              string                  `json:"budget"`
	TargetChainDecimals int                     `json:"targetChainDecimals"`
}

type ResultExecutionResponse struct {
	TransactionHash string `json:"transactionHash"`
	RefundStatus    string `json:"refundStatus"`
	RevertString    string `json:"revertString"`
	Status          string `json:"status"`
	GasUsed         string `json:"gasUsed"`
	Detail          string `json:"detail"`
}

type InstructionsResponse struct {
	EncodedExecutionInfo   string `json:"encodedExecutionInfo"`
	RefundAddress          string `json:"refundAddress"`
	SourceDeliveryProvider string `json:"sourceDeliveryProvider"`
	SenderAddress          string `json:"senderAddress"`
	VaaKeys                []any  `json:"vaaKeys"`
	ExtraReceiverValue     struct {
		Hex         string `json:"_hex"`
		IsBigNumber bool   `json:"_isBigNumber"`
	} `json:"extraReceiverValue"`
	TargetAddress          string `json:"targetAddress"`
	RequestedReceiverValue struct {
		Hex         string `json:"_hex"`
		IsBigNumber bool   `json:"_isBigNumber"`
	} `json:"requestedReceiverValue"`
	RefundChainID          int    `json:"refundChainId"`
	RefundDeliveryProvider string `json:"refundDeliveryProvider"`
	TargetChainID          int    `json:"targetChainId"`
}
