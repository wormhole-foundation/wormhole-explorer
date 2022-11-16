package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type VAAPushFunc func(context.Context, *vaa.VAA, []byte) error
