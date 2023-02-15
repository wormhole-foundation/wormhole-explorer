package health

import "context"

type HealthCheck struct {
	Status string `json:"status"`
}

type ReadyCheck struct {
	Ready string `json:"ready"`
}

type Check func(context.Context) error
