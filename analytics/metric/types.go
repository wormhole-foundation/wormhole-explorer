package metric

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Params struct {
	Source      string
	TrackID     string
	Vaa         *vaa.VAA
	VaaIsSigned bool
}

// MetricPushFunc is a function to push metrics
type MetricPushFunc func(context.Context, *Params) error
