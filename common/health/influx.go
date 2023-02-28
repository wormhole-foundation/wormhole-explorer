package health

import (
	"context"
	"errors"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	InfluxStatusOK = "pass"
)

// InfluxOSS do a healtheckIN influx OSS version.
func InfluxOSS(client influxdb2.Client) Check {
	return func(ctx context.Context) error {
		healthcheck, err := client.Health(ctx)
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
}

// Influx do a healtheck in influx Cloud version.
func Influx(client influxdb2.Client) Check {
	return func(ctx context.Context) error {
		check, err := client.Ping(ctx)
		if err != nil {
			return err
		}
		if !check {
			return errors.New("influx cloud not available")
		}
		return nil
	}
}
