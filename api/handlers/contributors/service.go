package contributors

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type Service struct {
	Contributors []string
	repo         *repository
	logger       *zap.Logger
}

func (s *Service) GetContributorsTotalValues(ctx context.Context) ([]ContributorTotalValuesDTO, error) {

	wg := sync.WaitGroup{}
	wg.Add(len(s.Contributors))
	results := make(chan ContributorTotalValuesDTO, len(s.Contributors))

	for i := range s.Contributors {
		go func(j int) {
			defer wg.Done()

			rowStats, err := s.repo.getContributorStats(ctx, s.Contributors[i])
			if err != nil {
				s.logger.Error("error fetching contributor stats", zap.Error(err), zap.String("contributor", s.Contributors[i]))
				return
			}
			activity, err := s.repo.getContributorActivity(ctx, s.Contributors[i])
			if err != nil {
				s.logger.Error("error fetching contributor activity", zap.Error(err), zap.String("contributor", s.Contributors[i]))
				return
			}

			totalMessagesNow, _ := strconv.ParseUint(rowStats.Latest.TotalMessages, 10, 64)
			totalMessagesAsFromLast24hr, _ := strconv.ParseUint(rowStats.Last24.TotalMessages, 10, 64)
			last24HrMessages := totalMessagesNow - totalMessagesAsFromLast24hr
			dto := ContributorTotalValuesDTO{
				Contributor:           s.Contributors[i],
				TotalValueLocked:      rowStats.Latest.TotalValueLocked,
				TotalMessages:         rowStats.Latest.TotalMessages,
				LastDayMessages:       strconv.FormatUint(last24HrMessages, 10),
				LastDayDiffPercentage: strconv.FormatFloat(float64(last24HrMessages/totalMessagesAsFromLast24hr)*100, 'f', 2, 64),
				TotalValueTransferred: activity.TotalValueTransferred,
				TotalValueSecured:     activity.TotalVolumeSecure,
			}
			results <- dto
		}(i)
	}

	wg.Wait()
	close(results)
	resultsSlice := make([]ContributorTotalValuesDTO, 0, len(s.Contributors))
	for r := range results {
		resultsSlice = append(resultsSlice, r)
	}
	if len(resultsSlice) == 0 {
		return nil, errors.New("failed fetching contributors total stats")
	}
	return resultsSlice, nil
}
