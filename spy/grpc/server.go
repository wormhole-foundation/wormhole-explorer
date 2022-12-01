package grpc

import (
	"fmt"
	"net"

	"github.com/certusone/wormhole/node/pkg/common"
	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	Runnable supervisor.Runnable
	srv      *grpc.Server
}

// NewServer creates a GRPC server.
func NewServer(h *Handler, logger *zap.Logger, listenAddr string) (*Server, error) {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	logger.Info("spy server listening", zap.String("addr", l.Addr().String()))

	grpcServer := common.NewInstrumentedGRPCServer(logger)
	spyv1.RegisterSpyRPCServiceServer(grpcServer, h)

	runnale := supervisor.GRPCServer(grpcServer, l, false)
	return &Server{Runnable: runnale, srv: grpcServer}, nil
}

// Stop stops the GRPC server gracefully.
func (s *Server) Stop() {
	s.srv.GracefulStop()
}
