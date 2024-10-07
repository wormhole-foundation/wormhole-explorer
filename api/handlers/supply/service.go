package supply

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
}

func NewService(logger *zap.Logger) *Service {
	return &Service{
		logger: logger.With(zap.String("module", "SupplyService")),
	}
}

func (s *Service) getCirculatingSupplyForDate(_ context.Context, date time.Time) Supply {
	firstUnlock := Supplies[0]
	lastUnlock := Supplies[len(Supplies)-1]
	daysDifference := int(date.Sub(firstUnlock.Day).Hours() / 24) // Calculate the difference in days

	if daysDifference < 0 {
		// If the date is before the initial date, return 0
		return Supply{}
	} else if daysDifference >= len(Supplies) {
		// If the date is beyond the last date, return the last unlock
		return lastUnlock
	}

	// Return the unlocked amount for the calculated index
	return Supplies[daysDifference]
}

func (s *Service) GetCurrentCirculatingSupply(ctx context.Context) Supply {
	return s.getCirculatingSupplyForDate(ctx, time.Now())
}

func (s *Service) GetTotalSupply(_ context.Context) int {
	return TotalSupply
}
