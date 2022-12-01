package grpc

import (
	"fmt"

	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler represents a GRPC subscription service handler.
type Handler struct {
	spyv1.UnimplementedSpyRPCServiceServer
	svs *SignedVaaSubscribers
	avs *AllVaaSubscribers
}

// NewHandler creates a new handler of suscriptions.
func NewHandler(svs *SignedVaaSubscribers, avs *AllVaaSubscribers) *Handler {
	return &Handler{
		svs: svs,
		avs: avs,
	}
}

// SubscribeSignedVAA implements the suscriptions of signed VAA.
func (h *Handler) SubscribeSignedVAA(req *spyv1.SubscribeSignedVAARequest, resp spyv1.SpyRPCService_SubscribeSignedVAAServer) error {
	var fi []filterSignedVaa
	if req.Filters != nil {
		for _, f := range req.Filters {
			switch t := f.Filter.(type) {
			case *spyv1.FilterEntry_EmitterFilter:
				addr, err := vaa.StringToAddress(t.EmitterFilter.EmitterAddress)
				if err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode emitter address: %v", err))
				}
				fi = append(fi, filterSignedVaa{
					chainId:     vaa.ChainID(t.EmitterFilter.ChainId),
					emitterAddr: addr,
				})
			default:
				return status.Error(codes.InvalidArgument, "unsupported filter type")
			}
		}
	}

	id, sub := h.svs.Register(fi)
	defer h.svs.Unregister(id)

	for {
		select {
		case <-resp.Context().Done():
			return resp.Context().Err()
		case msg := <-sub.ch:
			if err := resp.Send(&spyv1.SubscribeSignedVAAResponse{
				VaaBytes: msg.vaaBytes,
			}); err != nil {
				return err
			}
		}
	}
}

// SubscribeSignedVAAByType implements the suscriptions of signed VAA by type.
func (h *Handler) SubscribeSignedVAAByType(req *spyv1.SubscribeSignedVAAByTypeRequest, resp spyv1.SpyRPCService_SubscribeSignedVAAByTypeServer) error {
	var fi []*spyv1.FilterEntry
	if req.Filters != nil {
		for _, f := range req.Filters {
			switch t := f.Filter.(type) {

			case *spyv1.FilterEntry_EmitterFilter:
				// validate the emitter address is valid by decoding it
				_, err := vaa.StringToAddress(t.EmitterFilter.EmitterAddress)
				if err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode emitter address: %v", err))
				}
				fi = append(fi, &spyv1.FilterEntry{Filter: t})

			case *spyv1.FilterEntry_BatchFilter,
				*spyv1.FilterEntry_BatchTransactionFilter:
				fi = append(fi, &spyv1.FilterEntry{Filter: t})
			default:
				return status.Error(codes.InvalidArgument, "unsupported filter type")
			}
		}
	}

	id, sub := h.avs.Register(fi)
	defer h.avs.Unregister(id)

	for {
		select {
		case <-resp.Context().Done():
			return resp.Context().Err()
		case msg := <-sub.ch:
			if err := resp.Send(msg); err != nil {
				return err
			}
		}
	}
}
