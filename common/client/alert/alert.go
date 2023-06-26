package alert

import (
	"fmt"

	opsgenieAlert "github.com/opsgenie/opsgenie-go-sdk-v2/alert"
)

// Respoder struct to define alert routing notifications.
type Responder struct {
	Id       string
	Type     string
	Name     string
	Username string
}

// Priority is the alert priority.
type Priority string

const (
	CRITICAL      Priority = "CRITICAL"
	HIGH          Priority = "HIGH"
	MODERATE      Priority = "MODERATE"
	LOW           Priority = "LOW"
	INFORMATIONAL Priority = "INFORMATIONAL"
)

// Alert is the alert struct.
type Alert struct {
	Message     string
	Alias       string
	Description string
	Actions     []string
	Tags        []string
	Entity      string
	Priority    Priority
	Responder   []Responder
	VisibleTo   []Responder
	context     AlertContext
}

// AlertContext contains the alert execution context.
type AlertContext struct {
	Details map[string]string
	Error   error
	Note    string
}

// toOpsgenieResponder converts a Responder to an Opsgenie Responder.
func (a *Responder) toOpsgenieResponder() opsgenieAlert.Responder {

	var responderType opsgenieAlert.ResponderType
	switch a.Type {
	case "user":
		responderType = opsgenieAlert.UserResponder
	case "team":
		responderType = opsgenieAlert.TeamResponder
	case "escalation":
		responderType = opsgenieAlert.EscalationResponder
	case "schedule":
		responderType = opsgenieAlert.ScheduleResponder
	}

	opsgenieResponder := opsgenieAlert.Responder{
		Id:       a.Id,
		Type:     responderType,
		Name:     a.Name,
		Username: a.Username,
	}
	return opsgenieResponder
}

// toOpsgeniePriority converts a Priority to an Opsgenie Priority.
func (p Priority) toOpsgeniePriority() opsgenieAlert.Priority {
	switch p {
	case CRITICAL:
		return opsgenieAlert.P1
	case HIGH:
		return opsgenieAlert.P2
	case MODERATE:
		return opsgenieAlert.P3
	case LOW:
		return opsgenieAlert.P4
	case INFORMATIONAL:
		return opsgenieAlert.P5
	default:
		return opsgenieAlert.P5
	}
}

// toOpsgenieRequest converts an Alert to an Opsgenie CreateAlertRequest.
func (a Alert) toOpsgenieRequest() opsgenieAlert.CreateAlertRequest {
	// convert priority to opsgenie priority.
	priotity := a.Priority.toOpsgeniePriority()

	// convert responders to opsgenie responders.
	var responders []opsgenieAlert.Responder
	for _, responder := range a.Responder {
		responders = append(responders, responder.toOpsgenieResponder())
	}

	// convert visibleTo to opsgenie responders.
	var visibleTo []opsgenieAlert.Responder
	for _, responder := range a.VisibleTo {
		visibleTo = append(visibleTo, responder.toOpsgenieResponder())
	}

	// add error details to opsgenie alert details data.
	description := a.Description
	if a.context.Error != nil {
		description = fmt.Sprintf("%s\n%s", a.Description, a.context.Error.Error())
	}

	return opsgenieAlert.CreateAlertRequest{
		Message:     a.Message,
		Alias:       a.Alias,
		Description: description,
		Actions:     a.Actions,
		Tags:        a.Tags,
		Details:     a.context.Details,
		Entity:      a.Entity,
		Priority:    priotity,
		Note:        a.context.Note,
		Responders:  responders,
		VisibleTo:   visibleTo,
	}
}

// GetMessagePrefix returns the alert message prefix.
func GetMessagePrefix(enviroment, p2pPNetwork string) string {
	if enviroment == "production" {
		return fmt.Sprintf("[%s-%s]", enviroment, p2pPNetwork)
	}
	return fmt.Sprintf("[%s]", enviroment)
}
