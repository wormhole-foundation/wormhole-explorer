package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	ErrorSaveVAA            = "ERROR_SAVE_VAA"
	ErrorSavePyth           = "ERROR_SAVE_PYTH"
	ErrorSaveObservation    = "ERROR_SAVE_OBSERVATION"
	ErrorSaveHeartbeat      = "ERROR_SAVE_HEARTBEAT"
	ErrorSaveGovernorStatus = "ERROR_SAVE_GOVERNOR_STATUS"
	EroorSaveGovernorConfig = "ERROR_SAVE_GOVERNOR_CONFIG"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)

	messagePrefix := alert.GetMessagePrefix(cfg.Enviroment, cfg.P2PNetwork)
	// Alert error saving vaa.
	alerts[ErrorSaveVAA] = alert.Alert{
		Alias:       ErrorSaveVAA,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving VAA in vaas collection"),
		Description: "An error was found persisting the vaa in mongo in the vaas collection.",
		Actions:     []string{"check vaas collection, vaa may have persisted by retry", "check dead letter queue"},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "vaa", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}
	// Alert error saving pyth
	alerts[ErrorSavePyth] = alert.Alert{
		Alias:       ErrorSavePyth,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving pyth in vaasPythnet collection"),
		Description: "An error was found persisting the pyth in mongo in the vaasPythnet collection.",
		Actions:     []string{"pyth may have persisted by retry"},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "vaasPythnet", "mongo"},
		Entity:      "fly",
		Priority:    alert.INFORMATIONAL,
	}
	// Alert error saving observation
	alerts[ErrorSaveObservation] = alert.Alert{
		Alias:       ErrorSaveObservation,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving observation in observations collection"),
		Description: "An error was found persisting the observation in mongo in the observations collection.",
		Actions:     []string{},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "observations", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}
	// Alert error saving heartbeat
	alerts[ErrorSaveHeartbeat] = alert.Alert{
		Alias:       ErrorSaveHeartbeat,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving heartbeat in heartbeats collection"),
		Description: "An error was found persisting the heartbeat in mongo in the heartbeats collection.",
		Actions:     []string{},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "heartbeats", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}
	alerts[ErrorSaveGovernorStatus] = alert.Alert{
		Alias:       ErrorSaveGovernorStatus,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving governor status in governorStatus collection"),
		Description: "An error was found persisting the governor status in mongo in the governorStatus collection.",
		Actions:     []string{},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "governorStatus", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}
	alerts[EroorSaveGovernorConfig] = alert.Alert{
		Alias:       EroorSaveGovernorConfig,
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error saving governor config in governorConfig collection"),
		Description: "An error was found persisting the governor config in mongo in the governorConfig collection.",
		Actions:     []string{},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "governorConfig", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}

	return alerts
}
