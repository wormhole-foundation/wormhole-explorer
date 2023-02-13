package health

import (
	"context"
	"time"
)

// GuardianCheck definition.
type GuardianCheck struct {
	maxHealthTimeDuration time.Duration
	lastPing              time.Time
}

// NewGuardianCheck instanciate a new GuardianCheck
func NewGuardianCheck(maxHealthTimeSeconds int64) *GuardianCheck {
	return &GuardianCheck{maxHealthTimeDuration: time.Duration(maxHealthTimeSeconds), lastPing: time.Now()}
}

// Change last ping.
func (g *GuardianCheck) Ping(ctx context.Context) {
	g.lastPing = time.Now()
}

// IsAlive check if the guardians are alive.
func (g *GuardianCheck) IsAlive() bool {
	healthTime := time.Now().Add(-time.Second * 60)
	return !g.lastPing.Before(healthTime)
}
