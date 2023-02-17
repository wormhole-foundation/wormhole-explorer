package influx

import (
	"context"
	"errors"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	InfluxStatusOK = "pass"
)

type Client struct {
	influxCli influxdb2.Client
}

// NewClient create a new influxdb Client.
func NewClient(url, token string) *Client {
	return &Client{influxCli: influxdb2.NewClient(url, token)}
}

// Healthcheck healthcheck influx api.
func (f *Client) Healthcheck(ctx context.Context) error {
	healthcheck, err := f.influxCli.Health(ctx)
	if err != nil {
		return err
	}
	if healthcheck == nil {
		return errors.New("influx healthcheck can not be nil")
	}
	if healthcheck.Status != InfluxStatusOK {
		return fmt.Errorf("influx healthcheck status: %s", healthcheck.Status)
	}
	return nil
}
