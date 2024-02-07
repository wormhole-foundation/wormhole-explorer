package builder

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	healthcheck "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	flyAlert "github.com/wormhole-foundation/wormhole-explorer/fly/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
)

func NewAlertClient(cfg *config.Configuration) (alert.AlertClient, error) {
	if !cfg.AlertEnabled {
		return alert.NewDummyClient(), nil
	}
	alertConfig := alert.AlertConfig{
		Environment: cfg.Environment,
		Enabled:     cfg.AlertEnabled,
		ApiKey:      cfg.AlertApiKey,
	}
	return alert.NewAlertService(alertConfig, flyAlert.LoadAlerts)
}

func NewMetrics(cfg *config.Configuration) metrics.Metrics {
	if !cfg.MetricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.Environment)
}

func CheckGuardian(guardian *health.GuardianCheck) healthcheck.Check {
	return func(ctx context.Context) error {
		isAlive := guardian.IsAlive()
		if !isAlive {
			return errors.New("guardian healthcheck not arrive in time")
		}
		return nil
	}
}
