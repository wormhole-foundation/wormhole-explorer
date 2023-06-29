package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	ErrorDecodeWatcherEvent = "ERROR_DECODE_WATCHER_EVENT"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)
	messagePrefix := alert.GetMessagePrefix(cfg.Enviroment, cfg.P2PNetwork)

	// Alert error devoding watcher event.
	alerts[ErrorDecodeWatcherEvent] = alert.Alert{
		Alias:       "Error decoding watcher event",
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error decoding watcher event"),
		Description: "An error was found decoding the watcher event.",
		Actions:     []string{""},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "pipeline", "watcher", "mongo"},
		Entity:      "pipeline",
		Priority:    alert.CRITICAL,
	}

	return alerts
}
