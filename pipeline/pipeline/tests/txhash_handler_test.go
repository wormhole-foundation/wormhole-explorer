package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/pipeline"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/pipeline/mocks"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewTxHashHandler(t *testing.T) {

	mock := gomock.NewController(t)
	defer mock.Finish()

	repo := mocks.NewMockIRepository(mock)

	//	log, _ := zap.NewDevelopment()
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	quit := make(chan bool)

	var f = topic.PushFunc(func(context.Context, *topic.Event) error {
		return nil
	})

	txHashHandler := pipeline.NewTxHashHandler(repo, f, alert.NewDummyClient(), metrics.NewDummyMetrics(), observedLogger, quit)
	txHashHandler.AddVaaFixItem(topic.Event{
		ID: "vaa1",
	},
	)

	ctx := context.Background()

	repo.EXPECT().GetVaaIdTxHash(ctx, "vaa1").Return(nil, fmt.Errorf("error"))
	repo.EXPECT().GetVaaIdTxHash(ctx, "vaa1").Return(&pipeline.VaaIdTxHash{
		ChainID: 1,
		TxHash:  "0xbabla",
	}, nil)
	repo.EXPECT().UpdateVaaDocTxHash(ctx, "vaa1", "0xbabla").Return(nil)

	go txHashHandler.Run(ctx)
	time.Sleep(6 * time.Second)
	close(quit)

	require.Equal(t, 3, observedLogs.Len())
	allLogs := observedLogs.All()
	// first attempt to get txhash should fail
	assert.Equal(t, "Error while trying to fix vaa txhash", allLogs[1].Message)
	// second attempt to get txhash should succeed
	assert.Equal(t, "Vaa txhash fixed", allLogs[2].Message)

}
