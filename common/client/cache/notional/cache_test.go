package notional

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotionalCache_renderKey(t *testing.T) {

	nc := &NotionalCache{
		client: nil,
		prefix: "staging-mainnet",
	}

	key := nc.renderKey("BTC/USDT")
	assert.Equal(t, "staging-mainnet:BTC/USDT", key)

}

func TestNotionalCache_renderRegexp(t *testing.T) {

	nc := &NotionalCache{
		client: nil,
		prefix: "staging-mainnet",
	}

	key := nc.renderRegExp()
	assert.Equal(t, "*staging-mainnet:WORMSCAN:NOTIONAL:SYMBOL:*", key)

	nc = &NotionalCache{
		client: nil,
		prefix: "",
	}
	key = nc.renderRegExp()
	assert.Equal(t, "*WORMSCAN:NOTIONAL:SYMBOL:*", key)

}
