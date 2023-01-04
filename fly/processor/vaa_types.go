package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VAAPushFunc is a function to push VAA message.
type VAAPushFunc func(context.Context, *vaa.VAA, []byte) error

// VAANotifyFunc is a function to notify saved VAA message.
type VAANotifyFunc func(context.Context, *vaa.VAA, []byte) error
