package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func createGRPCServer(handler *Handler, logger *zap.Logger) (context.Context, *grpc.ClientConn, spyv1.SpyRPCServiceClient) {
	listen := bufconn.Listen(1024 * 1024)
	grpcServer := common.NewInstrumentedGRPCServer(logger, common.GrpcLogDetailMinimal)
	spyv1.RegisterSpyRPCServiceServer(grpcServer, handler)
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			logger.Fatal("Server exited with error", zap.Error(err))
		}
	}()
	ctx := context.Background()
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) {
				return listen.Dial()
			}), creds)
	if err != nil {
		logger.Fatal("Failed to dial bufnet", zap.Error(err))
	}

	client := spyv1.NewSpyRPCServiceClient(conn)
	return ctx, conn, client
}

func TestSubscribeSignedVAA_OK(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svs := NewSignedVaaSubscribers(logger)
	avs := NewAllVaaSubscribers(logger)
	handler := NewHandler(svs, avs, logger)

	_, _, client := createGRPCServer(handler, logger)

	t.Run("receive valid vaa", func(t *testing.T) {
		doneSvs := make(chan bool)
		ctx, cancel := context.WithCancel(context.TODO())
		go func(ctx context.Context) {
			defer close(doneSvs)
			svs.Start(ctx)
		}(ctx)
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		req := &spyv1.SubscribeSignedVAARequest{}
		stream, err := client.SubscribeSignedVAA(ctx, req)
		assert.Nil(t, err)
		doneCh := make(chan bool)
		go func() {
			defer close(doneCh)
			signedVAA, err := stream.Recv()
			assert.Nil(t, err)
			assert.NotNil(t, signedVAA)
			assert.Equal(t, vaaBytes, signedVAA.VaaBytes)

		}()
		waitForSignedSubscription(handler)
		err = svs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		<-doneCh
		cancel()
		<-doneSvs
	})
}

func TestSubscribeSignedVAA_Failed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svs := NewSignedVaaSubscribers(logger)
	avs := NewAllVaaSubscribers(logger)
	handler := NewHandler(svs, avs, logger)

	ctx, _, client := createGRPCServer(handler, logger)

	t.Run("invalid emitter address", func(t *testing.T) {
		req := &spyv1.SubscribeSignedVAARequest{
			Filters: []*spyv1.FilterEntry{
				{
					Filter: &spyv1.FilterEntry_EmitterFilter{
						EmitterFilter: &spyv1.EmitterFilter{
							ChainId:        publicrpcv1.ChainID_CHAIN_ID_ETHEREUM,
							EmitterAddress: "bad-address",
						},
					},
				},
			},
		}
		c, err := client.SubscribeSignedVAA(ctx, req)
		assert.Nil(t, err)
		_, err = c.Recv()
		assert.NotNil(t, err)
	})

	t.Run("unsupported filter type", func(t *testing.T) {
		req := &spyv1.SubscribeSignedVAARequest{
			Filters: []*spyv1.FilterEntry{
				{
					Filter: &spyv1.FilterEntry_BatchFilter{
						BatchFilter: &spyv1.BatchFilter{
							ChainId: publicrpcv1.ChainID_CHAIN_ID_ETHEREUM,
						},
					},
				},
			},
		}
		c, err := client.SubscribeSignedVAA(ctx, req)
		assert.Nil(t, err)
		_, err = c.Recv()
		assert.NotNil(t, err)
	})
}

func TestSubscribeSignedVAAByType_OK(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svs := NewSignedVaaSubscribers(logger)
	avs := NewAllVaaSubscribers(logger)
	handler := NewHandler(svs, avs, logger)

	_, _, client := createGRPCServer(handler, logger)

	t.Run("receive valid vaa", func(t *testing.T) {
		doneAvs := make(chan bool)
		ctx, cancel := context.WithCancel(context.TODO())
		go func(ctx context.Context) {
			defer close(doneAvs)
			avs.Start(ctx)
		}(ctx)
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		req := &spyv1.SubscribeSignedVAAByTypeRequest{}
		stream, err := client.SubscribeSignedVAAByType(ctx, req)
		assert.Nil(t, err)
		doneCh := make(chan bool)
		go func() {
			defer close(doneCh)
			resp, err := stream.Recv()
			assert.Nil(t, err)
			assert.NotNil(t, resp)
			v, ok := resp.VaaType.(*spyv1.SubscribeSignedVAAByTypeResponse_SignedVaa)
			assert.True(t, ok)
			assert.Equal(t, vaaBytes, v.SignedVaa.Vaa)
		}()
		waitForSignedVAAByTypeSubscription(handler)
		err = avs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		<-doneCh
		cancel()
		<-doneAvs
	})
}

func TestSubscribeSignedVAAByType_Failed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svs := NewSignedVaaSubscribers(logger)
	avs := NewAllVaaSubscribers(logger)
	handler := NewHandler(svs, avs, logger)

	ctx, _, client := createGRPCServer(handler, logger)

	t.Run("invalid emitter address", func(t *testing.T) {
		req := &spyv1.SubscribeSignedVAAByTypeRequest{
			Filters: []*spyv1.FilterEntry{
				{
					Filter: &spyv1.FilterEntry_EmitterFilter{
						EmitterFilter: &spyv1.EmitterFilter{
							ChainId:        publicrpcv1.ChainID_CHAIN_ID_ETHEREUM,
							EmitterAddress: "bad-address",
						},
					},
				},
			},
		}
		c, err := client.SubscribeSignedVAAByType(ctx, req)
		assert.Nil(t, err)
		_, err = c.Recv()
		assert.NotNil(t, err)
	})
}

func waitForSignedSubscription(handler *Handler) {
	tk := time.NewTicker(time.Millisecond * 100)
	for range tk.C {
		subs := len(handler.svs.subscribers)
		if subs > 0 {
			return
		}
	}
}

func waitForSignedVAAByTypeSubscription(handler *Handler) {
	tk := time.NewTicker(time.Millisecond * 100)
	for range tk.C {
		subs := len(handler.avs.subscribers)
		if subs > 0 {
			return
		}
	}
}
