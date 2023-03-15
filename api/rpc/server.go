package rpc

import (
	"github.com/certusone/wormhole/node/pkg/common"
	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	Srv *grpc.Server
}

// NewServer creates a GRPC server.
func NewServer(h *Handler, logger *zap.Logger) *grpc.Server {
	grpcServer := common.NewInstrumentedGRPCServer(logger, common.GrpcLogDetailMinimal)
	publicrpcv1.RegisterPublicRPCServiceServer(grpcServer, h)
	return grpcServer
}
