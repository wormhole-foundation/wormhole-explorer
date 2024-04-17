package grpc

import (
	"fmt"

	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler represents a GRPC subscription service handler.
type Handler struct {
	spyv1.UnimplementedSpyRPCServiceServer
	svs    *SignedVaaSubscribers
	logger *zap.Logger
}

// NewHandler creates a new handler of suscriptions.
func NewHandler(svs *SignedVaaSubscribers, logger *zap.Logger) *Handler {
	return &Handler{
		svs:    svs,
		logger: logger,
	}
}

// SubscribeSignedVAA implements the suscriptions of signed VAA.
func (h *Handler) SubscribeSignedVAA(req *spyv1.SubscribeSignedVAARequest, resp spyv1.SpyRPCService_SubscribeSignedVAAServer) error {
	h.logger.Info("Receiving new subscriber in signed VAA")
	var fi []filterSignedVaa
	if req.Filters != nil {
		for _, f := range req.Filters {
			switch t := f.Filter.(type) {
			case *spyv1.FilterEntry_EmitterFilter:
				addr, err := vaa.StringToAddress(t.EmitterFilter.EmitterAddress)
				if err != nil {
					h.logger.Error("Decoding emitter address", zap.Error(err))
					return status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode emitter address: %v", err))
				}
				fi = append(fi, filterSignedVaa{
					chainId:     vaa.ChainID(t.EmitterFilter.ChainId),
					emitterAddr: addr,
				})
			default:
				h.logger.Error("Unsupported filter type", zap.Any("filter", t))
				return status.Error(codes.InvalidArgument, "unsupported filter type")
			}
		}
	}

	subscriber := h.svs.Register(fi)
	defer h.svs.Unregister(subscriber)

	for {
		select {
		case <-resp.Context().Done():
			h.logger.Error("Context done", zap.String("id", subscriber.id), zap.Error(resp.Context().Err()))
			return resp.Context().Err()
		case msg := <-subscriber.ch:
			if err := resp.Send(&spyv1.SubscribeSignedVAAResponse{
				VaaBytes: msg.vaaBytes,
			}); err != nil {
				h.logger.Error("Sending vaas", zap.String("id", subscriber.id), zap.Error(err))
				return err
			}
		}
	}
}
