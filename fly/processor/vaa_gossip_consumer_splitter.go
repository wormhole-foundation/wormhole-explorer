package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// VAAGossipConsumerSplitterOption represents a consumer splitter option function.
type VAAGossipConsumerSplitterOption func(*VAAGossipConsumerSplitter)

// VAAGossipConsumerSplitter represents a vaa message splitter.
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

// NewVAAGossipSplitterConsumer creates a splitter instance.
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

// WithSize allows to specify channel size when setting a value.
func WithSize(v int) VAAGossipConsumerSplitterOption {
	return func(i *VAAGossipConsumerSplitter) {
		i.size = v
	}
}

// Push splits vaa message on different channels depending on whether it is a pyth or non pyth.
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

// Start runs two go routine to process messages for both channels.
func (p *VAAGossipConsumerSplitter) Start(ctx context.Context) {
	go p.executePyth(ctx)
	go p.executeNonPyth(ctx)
}

// Close closes all consumer resources.
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
