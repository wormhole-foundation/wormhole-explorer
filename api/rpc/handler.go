package rpc

import (
	"context"

	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler represents a GRPC subscription service handler.
type Handler struct {
	publicrpcv1.UnimplementedPublicRPCServiceServer
}

// NewHandler creates a new handler of suscriptions.
func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetLastHeartbeats(context.Context, *publicrpcv1.GetLastHeartbeatsRequest) (*publicrpcv1.GetLastHeartbeatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLastHeartbeats not implemented")
}
func (h *Handler) GetSignedVAA(context.Context, *publicrpcv1.GetSignedVAARequest) (*publicrpcv1.GetSignedVAAResponse, error) {
	return &publicrpcv1.GetSignedVAAResponse{VaaBytes: []byte{1, 2, 3}}, nil
}

func (h *Handler) GetSignedBatchVAA(context.Context, *publicrpcv1.GetSignedBatchVAARequest) (*publicrpcv1.GetSignedBatchVAAResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSignedBatchVAA not implemented")
}
func (h *Handler) GetCurrentGuardianSet(context.Context, *publicrpcv1.GetCurrentGuardianSetRequest) (*publicrpcv1.GetCurrentGuardianSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrentGuardianSet not implemented")
}
func (h *Handler) GovernorGetAvailableNotionalByChain(context.Context, *publicrpcv1.GovernorGetAvailableNotionalByChainRequest) (*publicrpcv1.GovernorGetAvailableNotionalByChainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GovernorGetAvailableNotionalByChain not implemented")
}
func (h *Handler) GovernorGetEnqueuedVAAs(context.Context, *publicrpcv1.GovernorGetEnqueuedVAAsRequest) (*publicrpcv1.GovernorGetEnqueuedVAAsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GovernorGetEnqueuedVAAs not implemented")
}
func (h *Handler) GovernorIsVAAEnqueued(context.Context, *publicrpcv1.GovernorIsVAAEnqueuedRequest) (*publicrpcv1.GovernorIsVAAEnqueuedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GovernorIsVAAEnqueued not implemented")
}
func (h *Handler) GovernorGetTokenList(context.Context, *publicrpcv1.GovernorGetTokenListRequest) (*publicrpcv1.GovernorGetTokenListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GovernorGetTokenList not implemented")
}
