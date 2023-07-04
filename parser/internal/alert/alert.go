package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	AlertKeyVaaPayloadParserError = "ERROR-CALL-VAA-PAYLOAD-PARSER"
	AlertKeyInsertParsedVaaError  = "ERROR-INSERT-PARSED-VAA"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)
	messagePrefix := alert.GetMessagePrefix(cfg.Enviroment, cfg.P2PNetwork)

	// Alert for VAA payload parser error.
	alerts[AlertKeyVaaPayloadParserError] = alert.Alert{
		Alias:       "Error calling VAA payload parser",
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error calling VAA payload parser"),
		Description: "An error was found calling VAA payload parser",
		Actions:     []string{""},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "parser", "vaa_payload_parser", "client"},
		Entity:      "parser",
		Priority:    alert.CRITICAL,
	}

	// Alert for insert parsed VAA error.
	alerts[AlertKeyInsertParsedVaaError] = alert.Alert{
		Alias:       "Error inserting parsed VAA",
		Message:     fmt.Sprintf("%s %s", messagePrefix, "Error inserting parsed VAA"),
		Description: "An error was found inserting parsed VAA",
		Actions:     []string{""},
		Tags:        []string{cfg.Enviroment, cfg.P2PNetwork, "parser", "parsedVaa", "mongo"},
		Entity:      "parser",
		Priority:    alert.CRITICAL,
	}

	return alerts
}
