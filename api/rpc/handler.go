package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	vaaservice "github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler rpc handler.
type Handler struct {
	publicrpcv1.UnimplementedPublicRPCServiceServer
	gs          guardian.GuardianSet
	vaaSrv      *vaaservice.Service
	hbSrv       *heartbeats.Service
	govSrv      *governor.Service
	guardianSrv *guardian.Service
	logger      *zap.Logger
	usePostgres bool
}

// NewHandler create a new rpc Handler.
func NewHandler(vaaSrv *vaaservice.Service, hbSrv *heartbeats.Service, govSrv *governor.Service, guardianSrv *guardian.Service, logger *zap.Logger, usePostgres bool) *Handler {
	return &Handler{vaaSrv: vaaSrv, hbSrv: hbSrv, govSrv: govSrv, guardianSrv: guardianSrv, logger: logger, usePostgres: usePostgres}
}

// GetSignedVAA get signedVAA by chainID, address, sequence.
func (h *Handler) GetSignedVAA(ctx context.Context, request *publicrpcv1.GetSignedVAARequest) (*publicrpcv1.GetSignedVAAResponse, error) {
	// check and get chainID/address/sequence
	if request.MessageId == nil {
		return nil, status.Error(codes.InvalidArgument, "no message ID specified")
	}

	chainID := vaa.ChainID(request.MessageId.EmitterChain.Number())

	// This interface is not supported for PythNet messages because those VAAs are not stored in the database.
	if chainID == vaa.ChainIDPythNet {
		return nil, status.Error(codes.InvalidArgument, "not supported for PythNet")
	}

	address, err := hex.DecodeString(request.MessageId.EmitterAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode address from hex: %v", err))
	}
	if len(address) != 32 {
		return nil, status.Error(codes.InvalidArgument, "address must be 32 bytes")
	}

	addr, err := types.BytesToAddress(address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to decode address from bytes: %v", err))
	}

	sequence := strconv.FormatUint(request.MessageId.Sequence, 10)

	// get VAA by Id.
	vaa, err := h.vaaSrv.FindById(
		ctx,
		false, //TODO modifiy api migration.
		chainID,
		addr,
		sequence,
		false, /*includeParsedPayload*/
	)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "requested VAA not found in store")
		}
		h.logger.Error("failed to fetch VAA", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "internal server error")
	}

	// build GetSignedVAAResponse response.
	return &publicrpcv1.GetSignedVAAResponse{
		VaaBytes: vaa.Data.Vaa,
	}, nil
}

// GetSignedBatchVAA get signed batch VAA.
func (h *Handler) GetSignedBatchVAA(ctx context.Context, _ any) (any, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// GetLastHeartbeats get last heartbeats.
func (h *Handler) GetLastHeartbeats(ctx context.Context, request *publicrpcv1.GetLastHeartbeatsRequest) (*publicrpcv1.GetLastHeartbeatsResponse, error) {
	// check guardianSet exists.
	if len(h.gs.GstByIndex) == 0 {
		return nil, status.Error(codes.Unavailable, "guardian set not fetched from chain yet")
	}

	// get lasted guardianSet.
	guardianSet := h.gs.GetLatest()
	guardianAddresses := guardianSet.KeysAsHexStrings()

	// get last heartbeats by ids.
	heartbeats, err := h.hbSrv.GetHeartbeatsByIds(ctx, false, guardianAddresses)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	response := &publicrpcv1.GetLastHeartbeatsResponse{
		Entries: make([]*publicrpcv1.GetLastHeartbeatsResponse_Entry, 0),
	}

	for _, hb := range heartbeats {
		networkResponses := make([]*gossipv1.Heartbeat_Network, 0, len(hb.Networks))
		for _, network := range hb.Networks {
			networkResponse := gossipv1.Heartbeat_Network{
				Id:              uint32(network.ID),
				Height:          network.Height,
				ContractAddress: network.ContractAddress,
				ErrorCount:      uint64(network.ErrorCount), //TODO:check
			}
			networkResponses = append(networkResponses, &networkResponse)
		}

		rawHeartbeat := gossipv1.Heartbeat{
			Counter:       hb.Counter,
			NodeName:      hb.NodeName,
			Timestamp:     hb.Timestamp,
			Networks:      networkResponses,
			Version:       hb.Version,
			GuardianAddr:  hb.GuardianAddr,
			BootTimestamp: hb.BootTimestamp,
			Features:      hb.Features,
		}

		response.Entries = append(response.Entries, &publicrpcv1.GetLastHeartbeatsResponse_Entry{
			VerifiedGuardianAddr: hb.ID,
			P2PNodeAddr:          "",
			RawHeartbeat:         &rawHeartbeat,
		})
	}
	return response, nil
}

// GetCurrentGuardianSet get current guardian set.
func (h *Handler) GetCurrentGuardianSet(ctx context.Context, request *publicrpcv1.GetCurrentGuardianSetRequest) (*publicrpcv1.GetCurrentGuardianSetResponse, error) {
	// check guardianSet exists.
	if len(h.gs.GstByIndex) == 0 {
		return nil, status.Error(codes.Unavailable, "guardian set not fetched from chain yet")
	}
	// get lasted guardianSet.
	guardinSet := h.gs.GetLatest()

	// get guardian addresses.
	addresses := make([]string, len(guardinSet.Keys))
	for i, v := range guardinSet.Keys {
		addresses[i] = v.Hex()
	}

	return &publicrpcv1.GetCurrentGuardianSetResponse{
		GuardianSet: &publicrpcv1.GuardianSet{
			Index:     guardinSet.Index,
			Addresses: addresses,
		},
	}, nil
}

// GovernorGetAvailableNotionalByChain get availableNotional.
func (h *Handler) GovernorGetAvailableNotionalByChain(ctx context.Context, _ *publicrpcv1.GovernorGetAvailableNotionalByChainRequest) (*publicrpcv1.GovernorGetAvailableNotionalByChainResponse, error) {
	availableNotional, err := h.govSrv.GetAvailNotionByChain(ctx, h.usePostgres)
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

// GovernorGetEnqueuedVAAs get enqueuedVaa.
func (h *Handler) GovernorGetEnqueuedVAAs(ctx context.Context, _ *publicrpcv1.GovernorGetEnqueuedVAAsRequest) (*publicrpcv1.GovernorGetEnqueuedVAAsResponse, error) {
	enqueuedVaa, err := h.govSrv.GetEnqueuedVaas(ctx, h.usePostgres)
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

// GovernorIsVAAEnqueued check if a vaa is enqueued.
func (h *Handler) GovernorIsVAAEnqueued(ctx context.Context, request *publicrpcv1.GovernorIsVAAEnqueuedRequest) (*publicrpcv1.GovernorIsVAAEnqueuedResponse, error) {

	if request.MessageId == nil {
		return nil, status.Error(codes.InvalidArgument, "Parameters are required")
	}

	chainID := vaa.ChainID(request.MessageId.EmitterChain)

	emitterAddress, err := types.StringToAddress(request.MessageId.EmitterAddress, false /*acceptSolanaFormat*/)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid emitter address")
	}

	isEnqueued, err := h.govSrv.IsVaaEnqueued(ctx, chainID, emitterAddress, strconv.FormatUint(request.MessageId.Sequence, 10), h.usePostgres)
	if err != nil {
		return nil, err
	}

	return &publicrpcv1.GovernorIsVAAEnqueuedResponse{IsEnqueued: isEnqueued}, nil
}

// GovernorGetTokenList get governor token list.
func (h *Handler) GovernorGetTokenList(ctx context.Context, _ *publicrpcv1.GovernorGetTokenListRequest) (*publicrpcv1.GovernorGetTokenListResponse, error) {
	tokenList, err := h.govSrv.GetTokenList(ctx, h.usePostgres)
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
