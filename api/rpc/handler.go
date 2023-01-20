package rpc

import (
	"context"
	"strconv"

	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	publicrpcv1.UnimplementedPublicRPCServiceServer
	srv *governor.Service
}

func NewHandler(srv *governor.Service) *Handler {
	return &Handler{srv: srv}
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

func (h *Handler) GovernorGetAvailableNotionalByChain(ctx context.Context, _ *publicrpcv1.GovernorGetAvailableNotionalByChainRequest) (*publicrpcv1.GovernorGetAvailableNotionalByChainResponse, error) {
	availableNotional, err := h.srv.GetAvailNotionByChain(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]*publicrpcv1.GovernorGetAvailableNotionalByChainResponse_Entry, 0)
	for _, v := range availableNotional {
		entry := publicrpcv1.GovernorGetAvailableNotionalByChainResponse_Entry{
			ChainId:                    uint32(v.ChainID),
			NotionalLimit:              uint64(v.NotionalLimit),
			RemainingAvailableNotional: uint64(v.AvailableNotional),
			BigTransactionSize:         uint64(v.MaxTransactionSize),
		}
		entries = append(entries, &entry)
	}
	response := publicrpcv1.GovernorGetAvailableNotionalByChainResponse{
		Entries: entries,
	}
	return &response, nil
}

func (h *Handler) GovernorGetEnqueuedVAAs(ctx context.Context, _ *publicrpcv1.GovernorGetEnqueuedVAAsRequest) (*publicrpcv1.GovernorGetEnqueuedVAAsResponse, error) {
	enqueuedVaa, err := h.srv.GetEnqueuedVaas(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]*publicrpcv1.GovernorGetEnqueuedVAAsResponse_Entry, 0, len(enqueuedVaa))
	for _, v := range enqueuedVaa {
		seqUint64, err := strconv.ParseUint(v.Sequence, 10, 64)
		if err != nil {
			return nil, err
		}
		entry := publicrpcv1.GovernorGetEnqueuedVAAsResponse_Entry{
			EmitterChain:   uint32(v.EmitterChain),
			EmitterAddress: v.EmitterAddress,
			Sequence:       seqUint64,
			ReleaseTime:    uint32(v.ReleaseTime),
			NotionalValue:  uint64(v.NotionalValue),
			TxHash:         v.TxHash,
		}
		entries = append(entries, &entry)
	}
	response := publicrpcv1.GovernorGetEnqueuedVAAsResponse{
		Entries: entries,
	}
	return &response, nil
}

func (h *Handler) GovernorIsVAAEnqueued(ctx context.Context, request *publicrpcv1.GovernorIsVAAEnqueuedRequest) (*publicrpcv1.GovernorIsVAAEnqueuedResponse, error) {
	if request.MessageId == nil {
		return nil, status.Error(codes.InvalidArgument, "Parameters are required")
	}
	chainID := vaa.ChainID(request.MessageId.EmitterChain)
	emitterAddress, err := vaa.StringToAddress(request.MessageId.EmitterAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid emitter address")
	}
	isEnqueued, err := h.srv.IsVaaEnqueued(ctx, chainID, emitterAddress, strconv.FormatUint(request.MessageId.Sequence, 10))
	if err != nil {
		return nil, err
	}
	return &publicrpcv1.GovernorIsVAAEnqueuedResponse{IsEnqueued: isEnqueued}, nil
}

func (h *Handler) GovernorGetTokenList(ctx context.Context, _ *publicrpcv1.GovernorGetTokenListRequest) (*publicrpcv1.GovernorGetTokenListResponse, error) {
	tokenList, err := h.srv.GetTokenList(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]*publicrpcv1.GovernorGetTokenListResponse_Entry, 0, len(tokenList))
	for _, t := range tokenList {
		entry := publicrpcv1.GovernorGetTokenListResponse_Entry{
			OriginChainId: uint32(t.OriginChainID),
			OriginAddress: t.OriginAddress,
			Price:         t.Price,
		}
		entries = append(entries, &entry)
	}

	response := publicrpcv1.GovernorGetTokenListResponse{
		Entries: entries,
	}

	return &response, nil
}
