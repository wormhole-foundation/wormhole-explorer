package processor

import (
	"context"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VAAPushFunc is a function to push VAA message.
type VAAPushFunc func(context.Context, *vaa.VAA, []byte) error

// VAANotifyFunc is a function to notify saved VAA message.
type VAANotifyFunc func(context.Context, *vaa.VAA, []byte) error

// VAAQueueConsumeFunc is a function to obtain messages from a queue
type VAAQueueConsumeFunc func(context.Context) <-chan queue.Message[[]byte]

// ObservationPushFunc is a function to push observation message.
type ObservationPushFunc func(ctx context.Context, o *gossipv1.SignedObservation) error

// VAAQueueConsumeFunc is a function to obtain messages from a queue
type ObservationQueueConsumeFunc func(context.Context) <-chan queue.Message[*gossipv1.SignedObservation]
