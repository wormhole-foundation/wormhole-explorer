package config

import (
	"github.com/wormhole-foundation/wormhole-explorer/common/settings"
)

// ServiceSettings models the configuration settings for the core-contract-watcher service.
type ServiceSettings struct {
	settings.Logger
	settings.MongoDB
	settings.Monitoring
}
