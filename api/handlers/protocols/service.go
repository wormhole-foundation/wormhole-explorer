package protocols

import (
	"context"
	"encoding/json"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	Protocols []string
	repo      *Repository
	logger    *zap.Logger
	cache     cache.Cache
	cacheKey  string
	cacheTTL  int
}

type ProtocolTotalValuesDTO struct {
	Protocol              string  `json:"protocol"`
	TotalMessages         uint64  `json:"total_messages"`
	TotalValueLocked      float64 `json:"total_value_locked,omitempty"`
	TotalValueSecured     float64 `json:"total_value_secured,omitempty"`
	TotalValueTransferred float64 `json:"total_value_transferred,omitempty"`
	LastDayMessages       uint64  `json:"last_day_messages,omitempty"`
	LastDayDiffPercentage string  `json:"last_day_diff_percentage,omitempty"`
	Error                 string  `json:"error,omitempty"`
}

func NewService(protocols []string, repo *Repository, logger *zap.Logger, cache cache.Cache, cacheKey string, cacheTTL int) *Service {
	return &Service{
		Protocols: protocols,
		repo:      repo,
		logger:    logger,
		cache:     cache,
		cacheKey:  cacheKey,
		cacheTTL:  cacheTTL,
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

func (s *Service) getProtocolTotalValues(ctx context.Context, wg *sync.WaitGroup, protocol string, results chan<- ProtocolTotalValuesDTO) {
	defer wg.Done()

	k := s.cacheKey + ":" + protocol
	cachedValue, errCache := s.cache.Get(ctx, k)
	if errCache == nil {
		var val ProtocolTotalValuesDTO
		errCacheUnmarshall := json.Unmarshal([]byte(cachedValue), &val)
		if errCacheUnmarshall == nil {
			results <- val
			return
		}
		s.logger.Error("error unmarshalling cache value", zap.Error(errCacheUnmarshall), zap.String("cache_key", k))
	}

	type statsResult struct {
		result stats
		Err    error
	}
	statsRes := make(chan statsResult, 1)
	go func() {
		rowStats, errStats := s.repo.getProtocolStats(ctx, protocol)
		statsRes <- statsResult{result: rowStats, Err: errStats}
		close(statsRes)
	}()

	activity, err := s.repo.getProtocolActivity(ctx, protocol)
	if err != nil {
		s.logger.Error("error fetching protocol activity", zap.Error(err), zap.String("protocol", protocol))
		results <- ProtocolTotalValuesDTO{Protocol: protocol, Error: err.Error()}
		return
	}

	rStats := <-statsRes
	if rStats.Err != nil {
		s.logger.Error("error fetching protocol stats", zap.Error(rStats.Err), zap.String("protocol", protocol))
		results <- ProtocolTotalValuesDTO{Protocol: protocol, Error: rStats.Err.Error()}
		return
	}

	dto := ProtocolTotalValuesDTO{
		Protocol:              protocol,
		TotalValueLocked:      rStats.result.Latest.TotalValueLocked,
		TotalMessages:         rStats.result.Latest.TotalMessages,
		TotalValueTransferred: activity.TotalValueTransferred,
		TotalValueSecured:     activity.TotalVolumeSecure,
	}

	totalMsgNow := rStats.result.Latest.TotalMessages
	totalMessagesAsFromLast24hr := rStats.result.Last24.TotalMessages
	if totalMessagesAsFromLast24hr != 0 {
		last24HrMessages := totalMsgNow - totalMessagesAsFromLast24hr
		dto.LastDayMessages = last24HrMessages
		dto.LastDayDiffPercentage = strconv.FormatFloat(float64(last24HrMessages)/float64(totalMessagesAsFromLast24hr)*100, 'f', 2, 64) + "%"
	}

	dtoJson, _ := json.Marshal(dto) // don't handle error since the full lifecycle of the dto is under this scope
	errCache = s.cache.Set(ctx, k, string(dtoJson), time.Duration(s.cacheTTL)*time.Minute)
	if errCache != nil {
		s.logger.Error("error setting cache", zap.Error(errCache), zap.String("cache_key", k))
	}

	results <- dto
}
