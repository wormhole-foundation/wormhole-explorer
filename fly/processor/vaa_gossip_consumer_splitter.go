package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VAAGossipConsumerSplitterOption func(*VAAGossipConsumerSplitter)

type VAAGossipConsumerSplitter struct {
	push      VAAPushFunc
	pythCh    chan *sppliterMessage
	nonPythCh chan *sppliterMessage
	logger    *zap.Logger
	size      int
}

type sppliterMessage struct {
	value *vaa.VAA
	data  []byte
}

func NewVAAGossipSplitterConsumer(
	publish VAAPushFunc,
	logger *zap.Logger,
	opts ...VAAGossipConsumerSplitterOption) *VAAGossipConsumerSplitter {
	v := &VAAGossipConsumerSplitter{
		push:   publish,
		logger: logger,
		size:   50,
	}
	for _, opt := range opts {
		opt(v)
	}
	v.pythCh = make(chan *sppliterMessage, v.size)
	v.nonPythCh = make(chan *sppliterMessage, v.size)
	return v
}

func WithSize(v int) VAAGossipConsumerSplitterOption {
	return func(i *VAAGossipConsumerSplitter) {
		i.size = v
	}
}

func (p *VAAGossipConsumerSplitter) Push(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error {
	msg := &sppliterMessage{
		value: v,
		data:  serializedVaa,
	}
	if vaa.ChainIDPythNet == v.EmitterChain {
		//if the pyth channel is full, deletes the oldest message and sends the new message
		select {
		case p.pythCh <- msg:
		default:
			select {
			case <-p.pythCh:
			default:
			}
			p.pythCh <- msg
		}
	} else {
		p.nonPythCh <- msg
	}
	return nil
}

func (p *VAAGossipConsumerSplitter) Start(ctx context.Context) {
	go p.executePyth(ctx)
	go p.executeNonPyth(ctx)
}

func (p *VAAGossipConsumerSplitter) Close() {
	close(p.nonPythCh)
	close(p.pythCh)
}

func (p *VAAGossipConsumerSplitter) executePyth(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, opened := <-p.pythCh:
			if !opened {
				return
			}
			_ = p.push(ctx, m.value, m.data)
		}
	}
}

func (p *VAAGossipConsumerSplitter) executeNonPyth(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, opened := <-p.nonPythCh:
			if !opened {
				return
			}
			_ = p.push(ctx, m.value, m.data)
		}
	}
}
