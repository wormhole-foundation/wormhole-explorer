package notifier

import (
	"testing"

	"github.com/test-go/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
)

func TestNewLastSequenceNotifier(t *testing.T) {

	l := NewLastSequenceNotifier(nil, "mainnet-staging")

	assert.Equal(t, "mainnet-staging:wormscan:vaa-max-sequence", l.prefix)
}

func TestNewLastSequenceNotifierBackwardsCompat(t *testing.T) {

	l := NewLastSequenceNotifier(nil, "")

	assert.Equal(t, "wormscan:vaa-max-sequence", l.prefix)
}

func TestNewLastSequenceNotifierWithPrefix(t *testing.T) {

	cfg := config.Configuration{
		Environment: "staging",
		P2pNetwork:  "mainnet",
	}
	prefix := cfg.GetPrefix()

	l := NewLastSequenceNotifier(nil, prefix)

	assert.Equal(t, "mainnet-staging:wormscan:vaa-max-sequence", l.prefix)
}
