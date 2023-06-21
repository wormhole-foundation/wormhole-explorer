package alert

import (
	"context"
	"errors"

	opsgenieAlert "github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

// RegisterAlertsFunc is the function that loads the alerts from the corresponding component.
type RegisterAlertsFunc func(cfg AlertConfig) map[string]Alert

type AlertClient interface {
	CreateAlert(key string, alertCtx AlertContext) (Alert, error)
	Send(ctx context.Context, alert Alert) error
	CreateAndSend(ctx context.Context, key string, alertCtx AlertContext) error
}

type AlertConfig struct {
	Enviroment string
	P2PNetwork string
	ApiKey     string
	Enabled    bool
	Responder  []Responder
	VisibleTo  []Responder
}

// OpsgenieClient is the alert client.
type OpsgenieClient struct {
	enabled bool
	client  *opsgenieAlert.Client
	alerts  map[string]Alert
}

// NewAlertService creates a new alert service
func NewAlertService(cfg AlertConfig, registerAlertsFunc RegisterAlertsFunc) (*OpsgenieClient, error) {
	// load the alert templates from the corresponding component
	alerts := registerAlertsFunc(cfg)

	// create the opsgenie alert client
	alertClient, err := opsgenieAlert.NewClient(&client.Config{ApiKey: cfg.ApiKey})
	if err != nil {
		return nil, err
	}

	return &OpsgenieClient{
		client:  alertClient,
		alerts:  alerts,
		enabled: cfg.Enabled}, nil
}

// CreateAlert creates an alert by key and alert context.
// The key is the alert name, and with it we can get the alert from the registerd alerts.
// The alert context contains the alert execution data
func (s *OpsgenieClient) CreateAlert(key string, alertCtx AlertContext) (Alert, error) {
	if !s.enabled {
		return Alert{}, errors.New("alert not enabled")
	}
	// check alert exists.
	alert, ok := s.alerts[key]
	if !ok {
		return Alert{}, errors.New("alert not found")
	}

	alert.context = alertCtx
	return alert, nil
}

// Send sends an alert to opsgenie.
func (s *OpsgenieClient) Send(ctx context.Context, alert Alert) error {
	if !s.enabled {
		return errors.New("alert not enabled")
	}

	// check alert exists
	if alert.Message == "" {
		return errors.New("message can not be empty")
	}

	// convert alert to an opsgenie alerte request.
	alertRequest := alert.toOpsgenieRequest()

	// create the request
	_, err := s.client.Create(ctx, &alertRequest)
	if err != nil {
		return err
	}
	return nil
}

// CreateAndSend creates an alert by key and alert context and sends it to opsgenie.
func (s *OpsgenieClient) CreateAndSend(ctx context.Context, key string, alertCtx AlertContext) error {
	alert, err := s.CreateAlert(key, alertCtx)
	if err != nil {
		return err
	}
	return s.Send(ctx, alert)
}
