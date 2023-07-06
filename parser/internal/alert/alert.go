package alert

import (
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
)

// alert key constants definition.
const (
	AlertKeyVaaPayloadParserError = "ERROR-REQUEST-VAA-PAYLOAD-PARSER"
	AlertKeyInsertParsedVaaError  = "ERROR-INSERT-PARSED-VAA"
)

func LoadAlerts(cfg alert.AlertConfig) map[string]alert.Alert {
	alerts := make(map[string]alert.Alert)

	// Alert for VAA payload parser error.
	alerts[AlertKeyVaaPayloadParserError] = alert.Alert{
		Alias:       "Error calling VAA payload parser",
		Message:     fmt.Sprintf("[%s] %s", cfg.Environment, "Error calling VAA payload parser"),
		Description: "An error was found calling VAA payload parser",
		Actions:     []string{""},
		Tags:        []string{cfg.Environment, "parser", "vaa_payload_parser", "client"},
		Entity:      "parser",
		Priority:    alert.CRITICAL,
	}

	// Alert for insert parsed VAA error.
	alerts[AlertKeyInsertParsedVaaError] = alert.Alert{
		Alias:       "Error inserting parsed VAA",
		Message:     fmt.Sprintf("[%s] %s", cfg.Environment, "Error inserting parsed VAA"),
		Description: "An error was found inserting parsed VAA",
		Actions:     []string{""},
		Tags:        []string{cfg.Environment, "parser", "parsedVaa", "mongo"},
		Entity:      "parser",
		Priority:    alert.CRITICAL,
	}

	return alerts
}
