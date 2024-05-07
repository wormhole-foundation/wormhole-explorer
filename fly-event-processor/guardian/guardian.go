package guardian

import (
	client "github.com/wormhole-foundation/wormhole-explorer/common/client/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
)

// GuardianAPIClient is a wrapper around the Guardian API client and the pool of providers.
type GuardianAPIClient struct {
	Client *client.GuardianAPIClient
	Pool   *pool.Pool
}

// NewGuardianAPIClient creates a new Guardian API client.
func NewGuardianAPIClient(client *client.GuardianAPIClient, pool *pool.Pool) *GuardianAPIClient {
	return &GuardianAPIClient{
		Client: client,
		Pool:   pool,
	}
}
