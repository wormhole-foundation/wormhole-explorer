package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	ErrorSaveVAA = "ERROR_SAVE_VAA"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)

	// Alert Error saving vaa.
	alerts[ErrorSaveVAA] = alert.Alert{
		Alias:       ErrorSaveVAA,
		Message:     fmt.Sprintf("[%s-%s] %s", cfg.Enviroment, cfg.P2PNetwork, "Error saving VAA in vaas collection"),
		Description: "An error was found persisting the vaa in mongo in the vaas collection.",
		Actions:     []string{"check vaas collection, vaa may have persisted by retry", "check dead letter queue"},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "fly", "vaa", "mongo"},
		Entity:      "fly",
		Priority:    alert.CRITICAL,
	}
	return alerts
}
