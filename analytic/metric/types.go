package metric

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// MetricPushFunc is a function to push metrics
type MetricPushFunc func(context.Context, *vaa.VAA) error
