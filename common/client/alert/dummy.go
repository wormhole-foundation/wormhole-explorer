package alert

import "context"

// DummyClient is a dummy alert client.
type DummyClient struct{}

// NewDummyClient creates a new dummy alert client.
func NewDummyClient() *DummyClient {
	return &DummyClient{}
}

// NewDummyClient creates a new dummy alert client.
func (d *DummyClient) CreateAlert(key string, alertCtx AlertContext) (Alert, error) {
	return Alert{}, nil
}

// Send sends an alert to opsgenie.
func (d *DummyClient) Send(ctx context.Context, alert Alert) error {
	return nil
}

// CreateAndSend creates an alert by key and alert context and sends it to opsgenie.
func (d *DummyClient) CreateAndSend(ctx context.Context, key string, alertCtx AlertContext) error {
	return nil
}
