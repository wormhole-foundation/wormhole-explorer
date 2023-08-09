package config

import (
	"github.com/wormhole-foundation/wormhole-explorer/common/settings"
)

// ServiceSettings models the configuration settings for the event-watcher service.
type ServiceSettings struct {
	settings.Logger
	settings.MongoDB
	settings.Monitoring
	settings.P2p
	WatcherSettings
}

type WatcherSettings struct {
	EthereumRequestsPerMinute uint   `split_words:"true" default:"INFO"`
	EthereumUrl               string `split_words:"true" default:"INFO"`
	EthereumAuth              string `split_words:"true" default:"INFO"`
}
