package config

import (
	"github.com/wormhole-foundation/wormhole-explorer/common/settings"
)

// ServiceSettings models the configuration settings for the event-watcher service.
type ServiceSettings struct {
	settings.Logger
	settings.MongoDB
	settings.Monitoring
}
