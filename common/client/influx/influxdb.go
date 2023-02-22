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

// HealthcheckOSS healthcheck influx OSS version.
func (i *Client) HealthcheckOSS(ctx context.Context) error {
	healthcheck, err := i.influxCli.Health(ctx)
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

// Healthcheck healthcheck influxdb cloud version.
func (i *Client) Healthcheck(ctx context.Context) error {
	check, err := i.influxCli.Ping(ctx)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("influx cloud not available")
	}
	return nil
}
