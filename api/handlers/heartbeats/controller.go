package heartbeats

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(srv *Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "HeartbeatsController")),
	}
}

// HeartbeatsResponse response.
type HeartbeatsResponse struct {
	Heartbeats []*HeartbeatResponse `json:"entries"`
}

type HeartbeatResponse struct {
	VerifiedGuardianAddr string        `json:"verifiedGuardianAddr"`
	P2PNodeAddr          string        `json:"p2pNodeAddr"`
	RawHeartbeat         *RawHeartbeat `json:"rawHeartbeat"`
}

type RawHeartbeat struct {
	NodeName      string                      `json:"nodeName"`
	Counter       int64                       `json:"counter"`
	Timestamp     string                      `json:"timestamp"`
	Networks      []*HeartbeatNetworkResponse `json:"networks"`
	Version       string                      `json:"version"`
	GuardianAddr  string                      `json:"guardianAddr"`
	BootTimestamp string                      `json:"bootTimestamp"`
	Features      []string                    `json:"features"`
}

// HeartbeatNetwork definition.
type HeartbeatNetworkResponse struct {
	ID              int64  `bson:"id" json:"id"`
	Height          string `bson:"height" json:"height"`
	ContractAddress string `bson:"contractaddress" json:"contractAddress"`
	ErrorCount      string `bson:"errorcount" json:"errorCount"`
}

// GetLastHeartbeats handler for the endpoint /guardian_public_api/v1/heartbeats
// This endpoint has been migrated from the guardian grpc api.
func (c *Controller) GetLastHeartbeats(ctx *fiber.Ctx) error {
	// check guardianSet exists.
	if len(guardian.ByIndex) == 0 {
		return response.NewApiError(ctx, fiber.StatusServiceUnavailable, response.Unavailable,
			"guardian set not fetched from chain yet", nil)
	}
	// get lasted guardianSet.
	guardianSet := guardian.GetLatest()
	guardianAddresses := guardianSet.KeysAsHexStrings()

	// get last heartbeats by ids.
	heartbeats, err := c.srv.GetHeartbeatsByIds(ctx.Context(), guardianAddresses)
	if err != nil {
		return err
	}

	// build heartbeats response compatible with grpc api response.
	response := buildHeartbeatResponse(heartbeats)
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func buildHeartbeatResponse(heartbeats []*HeartbeatDoc) *HeartbeatsResponse {
	if heartbeats == nil {
		return nil
	}
	heartbeatResponses := make([]*HeartbeatResponse, 0, len(heartbeats))
	for _, heartbeat := range heartbeats {

		networkResponses := make([]*HeartbeatNetworkResponse, 0, len(heartbeat.Networks))
		for _, network := range heartbeat.Networks {
			networkResponse := &HeartbeatNetworkResponse{
				ID:              network.ID,
				Height:          strconv.Itoa(int(network.Height)),
				ContractAddress: network.ContractAddress,
				ErrorCount:      strconv.Itoa(int(network.ErrorCount)),
			}
			networkResponses = append(networkResponses, networkResponse)
		}

		hr := HeartbeatResponse{
			VerifiedGuardianAddr: heartbeat.ID,
			P2PNodeAddr:          "", // not exists in heartbeats mongo collection.
			RawHeartbeat: &RawHeartbeat{
				NodeName:      heartbeat.NodeName,
				Counter:       heartbeat.Counter,
				Timestamp:     strconv.Itoa(int(heartbeat.Timestamp)),
				Networks:      networkResponses,
				Version:       heartbeat.Version,
				GuardianAddr:  heartbeat.GuardianAddr,
				BootTimestamp: strconv.Itoa(int(heartbeat.BootTimestamp)),
				Features:      heartbeat.Features,
			},
		}
		heartbeatResponses = append(heartbeatResponses, &hr)
	}
	return &HeartbeatsResponse{
		Heartbeats: heartbeatResponses,
	}
}
