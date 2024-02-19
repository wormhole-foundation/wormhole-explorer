package protocols

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type Service struct {
	Protocols []string
	repo      *Repository
	logger    *zap.Logger
}

type ProtocolTotalValuesDTO struct {
	Protocol              string `json:"protocol"`
	TotalValueLocked      string `json:"total_value_locked,omitempty"`
	TotalValueSecured     string `json:"total_value_secured,omitempty"`
	TotalValueTransferred string `json:"total_value_transferred,omitempty"`
	TotalMessages         string `json:"total_messages"`
	LastDayMessages       string `json:"last_day_messages"`
	LastDayDiffPercentage string `json:"last_day_diff_percentage"`
	Error                 string `json:"error,omitempty"`
}

func NewService(protocols []string, repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		Protocols: protocols,
		repo:      repo,
		logger:    logger,
	}
}

func (s *Service) GetProtocolsTotalValues(ctx context.Context) []ProtocolTotalValuesDTO {

	wg := &sync.WaitGroup{}
	wg.Add(len(s.Protocols))
	results := make(chan ProtocolTotalValuesDTO, len(s.Protocols))

	for i := range s.Protocols {
		go s.getProtocolTotalValues(ctx, wg, s.Protocols[i], results)
	}
	wg.Wait()
	close(results)

	resultsSlice := make([]ProtocolTotalValuesDTO, 0, len(s.Protocols))
	for r := range results {
		resultsSlice = append(resultsSlice, r)
	}
	return resultsSlice
}

func (s *Service) getProtocolTotalValues(ctx context.Context, wg *sync.WaitGroup, contributor string, results chan<- ProtocolTotalValuesDTO) {
	defer wg.Done()

	type statsResult struct {
		result stats
		Err    error
	}
	statsRes := make(chan statsResult, 1)
	go func() {
		rowStats, errStats := s.repo.getProtocolStats(ctx, contributor)
		statsRes <- statsResult{result: rowStats, Err: errStats}
		close(statsRes)
	}()

	activity, err := s.repo.getProtocolActivity(ctx, contributor)
	if err != nil {
		s.logger.Error("error fetching protocol activity", zap.Error(err), zap.String("protocol", contributor))
		results <- ProtocolTotalValuesDTO{Protocol: contributor, Error: err.Error()}
		return
	}

	rStats := <-statsRes
	if rStats.Err != nil {
		s.logger.Error("error fetching protocol stats", zap.Error(rStats.Err), zap.String("protocol", contributor))
		results <- ProtocolTotalValuesDTO{Protocol: contributor, Error: rStats.Err.Error()}
		return
	}

	totalMessagesNow, _ := strconv.ParseUint(rStats.result.Latest.TotalMessages, 10, 64)
	totalMessagesAsFromLast24hr, _ := strconv.ParseUint(rStats.result.Last24.TotalMessages, 10, 64)
	last24HrMessages := totalMessagesNow - totalMessagesAsFromLast24hr
	dto := ProtocolTotalValuesDTO{
		Protocol:              contributor,
		TotalValueLocked:      rStats.result.Latest.TotalValueLocked,
		TotalMessages:         rStats.result.Latest.TotalMessages,
		LastDayMessages:       strconv.FormatUint(last24HrMessages, 10),
		LastDayDiffPercentage: strconv.FormatFloat(float64(last24HrMessages)/float64(totalMessagesAsFromLast24hr)*100, 'f', 2, 64) + "%",
		TotalValueTransferred: activity.TotalValueTransferred,
		TotalValueSecured:     activity.TotalVolumeSecure,
	}
	results <- dto
}
