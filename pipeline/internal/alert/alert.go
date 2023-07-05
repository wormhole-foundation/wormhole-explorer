package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	ErrorDecodeWatcherEvent = "ERROR_DECODE_WATCHER_EVENT"
	ErrorUpdateVaaTxHash    = "ERROR_UPDATE_VAA_TX_HASH"
	ErrorPushEventSNS       = "ERROR_PUSH_EVENT_SNS"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)

	// Alert error devoding watcher event.
	alerts[ErrorDecodeWatcherEvent] = alert.Alert{
		Alias:       "Error decoding watcher event",
		Message:     fmt.Sprintf("[%s] %s", cfg.Environment, "Error decoding watcher event"),
		Description: "An error was found decoding the watcher event from the mongo stream",
		Actions:     []string{""},
		Tags:        []string{cfg.Environment, "pipeline", "watcher", "mongo"},
		Entity:      "pipeline",
		Priority:    alert.CRITICAL,
	}

	// Alert error updating vaa txhash.
	alerts[ErrorUpdateVaaTxHash] = alert.Alert{
		Alias:       "Error updating vaa txhash",
		Message:     fmt.Sprintf("[%s] %s", cfg.Environment, "Error updating vaa txhash"),
		Description: "An error was found updating the vaa txhash",
		Actions:     []string{""},
		Tags:        []string{cfg.Environment, "pipeline", "vaa", "txHash", "mongo"},
		Entity:      "pipeline",
		Priority:    alert.CRITICAL,
	}

	// Alert error pushing event.
	alerts[ErrorPushEventSNS] = alert.Alert{
		Alias:       "Error pushing event to sns",
		Message:     fmt.Sprintf("[%s] %s", cfg.Environment, "Error pushing event to sns"),
		Description: "An error was found pushing the event to sns",
		Actions:     []string{""},
		Tags:        []string{cfg.Environment, "pipeline", "push", "sns"},
		Entity:      "pipeline",
		Priority:    alert.CRITICAL,
	}

	return alerts
}
