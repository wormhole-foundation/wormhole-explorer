package processor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap/zaptest"
)

func TestVAAGossipConsumerSplitter_PushPyth(t *testing.T) {
	ctx := context.TODO()
	messagesProcessed := 0
	pushFunc := func(_ context.Context, v *vaa.VAA, d []byte) error {
		messagesProcessed++
		time.Sleep(1 * time.Second)
		return nil
	}
	logger := zaptest.NewLogger(t)
	splitter := NewVAAGossipSplitterConsumer(pushFunc, 1, logger, WithSize(1))
	splitter.Start(ctx)

	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDPythNet, Sequence: 1}, nil)
	time.Sleep(500 * time.Millisecond)
	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDPythNet, Sequence: 2}, nil)
	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDPythNet, Sequence: 3}, nil)

	time.Sleep(5 * time.Second)
	splitter.Close()
	assert.Equal(t, 2, messagesProcessed)
}

func TestVAAGossipConsumerSplitter_PushNonPyth(t *testing.T) {
	ctx := context.TODO()
	messagesProcessed := 0
	pushFunc := func(_ context.Context, v *vaa.VAA, d []byte) error {
		messagesProcessed++
		time.Sleep(1 * time.Second)
		return nil
	}
	logger := zaptest.NewLogger(t)
	splitter := NewVAAGossipSplitterConsumer(pushFunc, 1, logger, WithSize(1))
	splitter.Start(ctx)

	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDEthereum, Sequence: 1}, nil)
	time.Sleep(500 * time.Millisecond)
	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDSolana, Sequence: 1}, nil)
	splitter.Push(ctx, &vaa.VAA{EmitterChain: vaa.ChainIDAlgorand, Sequence: 1}, nil)

	time.Sleep(5 * time.Second)
	splitter.Close()
	assert.Equal(t, 3, messagesProcessed)
}
