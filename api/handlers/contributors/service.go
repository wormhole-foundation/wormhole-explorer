package contributors

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type Service struct {
	Contributors []string
	repo         *Repository
	logger       *zap.Logger
}

type ContributorTotalValuesDTO struct {
	Contributor           string `json:"contributor"`
	TotalValueLocked      string `json:"total_value_locked,omitempty"`
	TotalValueSecured     string `json:"total_value_secured,omitempty"`
	TotalValueTransferred string `json:"total_value_transferred,omitempty"`
	TotalMessages         string `json:"total_messages"`
	LastDayMessages       string `json:"last_day_messages"`
	LastDayDiffPercentage string `json:"last_day_diff_percentage"`
	Error                 error
}

func NewService(contributors []string, repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		Contributors: contributors,
		repo:         repo,
		logger:       logger,
	}
}

func (s *Service) GetContributorsTotalValues(ctx context.Context) []ContributorTotalValuesDTO {

	wg := &sync.WaitGroup{}
	wg.Add(len(s.Contributors))
	results := make(chan ContributorTotalValuesDTO, len(s.Contributors))

	for i := range s.Contributors {
		go s.getContributorTotalValues(ctx, wg, s.Contributors[i], results)
	}
	wg.Wait()
	close(results)

	resultsSlice := make([]ContributorTotalValuesDTO, 0, len(s.Contributors))
	for r := range results {
		resultsSlice = append(resultsSlice, r)
	}
	return resultsSlice
}

func (s *Service) getContributorTotalValues(ctx context.Context, wg *sync.WaitGroup, contributor string, results chan<- ContributorTotalValuesDTO) {
	defer wg.Done()

	type statsResult struct {
		result stats
		Err    error
	}
	statsRes := make(chan statsResult, 1)
	go func() {
		rowStats, errStats := s.repo.getContributorStats(ctx, contributor)
		statsRes <- statsResult{result: rowStats, Err: errStats}
		close(statsRes)
	}()

	activity, err := s.repo.getContributorActivity(ctx, contributor)
	if err != nil {
		s.logger.Error("error fetching contributor activity", zap.Error(err), zap.String("contributor", contributor))
		results <- ContributorTotalValuesDTO{Contributor: contributor, Error: err}
		return
	}

	rStats := <-statsRes
	if rStats.Err != nil {
		s.logger.Error("error fetching contributor stats", zap.Error(rStats.Err), zap.String("contributor", contributor))
		results <- ContributorTotalValuesDTO{Contributor: contributor, Error: rStats.Err}
		return
	}

	totalMessagesNow, _ := strconv.ParseUint(rStats.result.Latest.TotalMessages, 10, 64)
	totalMessagesAsFromLast24hr, _ := strconv.ParseUint(rStats.result.Last24.TotalMessages, 10, 64)
	last24HrMessages := totalMessagesNow - totalMessagesAsFromLast24hr
	dto := ContributorTotalValuesDTO{
		Contributor:           contributor,
		TotalValueLocked:      rStats.result.Latest.TotalValueLocked,
		TotalMessages:         rStats.result.Latest.TotalMessages,
		LastDayMessages:       strconv.FormatUint(last24HrMessages, 10),
		LastDayDiffPercentage: strconv.FormatFloat(float64(last24HrMessages)/float64(totalMessagesAsFromLast24hr)*100, 'f', 2, 64) + "%",
		TotalValueTransferred: activity.TotalValueTransferred,
		TotalValueSecured:     activity.TotalVolumeSecure,
	}
	results <- dto
}
