package protocols

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CCTP = "CCTP_WORMHOLE_INTEGRATION"
const PortalTokenBridge = "PORTAL_TOKEN_BRIDGE"
const NTT = "NATIVE_TOKEN_TRANSFER"

type Service struct {
	Protocols      []string
	repo           *Repository
	logger         *zap.Logger
	coreProtocols  []string
	cache          cache.Cache
	cacheKeyPrefix string
	cacheTTL       int
	metrics        metrics.Metrics
	tvl            tvlProvider
}

type ProtocolTotalValuesDTO struct {
	ProtocolStats
	Error string `json:"error,omitempty"`
}

type ProtocolStats struct {
	Protocol              string  `json:"protocol"`
	TotalMessages         uint64  `json:"total_messages"`
	TotalValueLocked      float64 `json:"total_value_locked,omitempty"`
	TotalValueSecured     float64 `json:"total_value_secured,omitempty"`
	TotalValueTransferred float64 `json:"total_value_transferred,omitempty"`
	LastDayMessages       uint64  `json:"last_day_messages,omitempty"`
	LastDayDiffPercentage string  `json:"last_day_diff_percentage,omitempty"`
}

type tvlProvider interface {
	Get(ctx context.Context) (string, error)
}

type fetchProtocolStats func(ctx context.Context, protocol string) (intStats, error)

func NewService(extProtocols, coreProtocols []string, repo *Repository, logger *zap.Logger, cache cache.Cache, cacheKeyPrefix string, cacheTTL int, metrics metrics.Metrics, tvlProvider tvlProvider) *Service {
	return &Service{
		Protocols:      extProtocols,
		repo:           repo,
		logger:         logger,
		coreProtocols:  coreProtocols,
		cache:          cache,
		cacheKeyPrefix: cacheKeyPrefix,
		cacheTTL:       cacheTTL,
		metrics:        metrics,
		tvl:            tvlProvider,
	}
}

func (s *Service) GetProtocolsTotalValues(ctx context.Context) []ProtocolTotalValuesDTO {

	totalProtocols := len(s.Protocols) + len(s.coreProtocols)
	results := make(chan ProtocolTotalValuesDTO, totalProtocols)
	wg := &sync.WaitGroup{}
	wg.Add(totalProtocols)

	for _, p := range s.Protocols {
		go s.fetchProtocolValues(ctx, wg, p, results, s.getProtocolStats)
	}
	for _, p := range s.coreProtocols {
		go s.fetchProtocolValues(ctx, wg, p, results, s.getCoreProtocolStats)
	}

	wg.Wait()
	close(results)

	resultsSlice := make([]ProtocolTotalValuesDTO, 0, len(s.Protocols))
	for r := range results {
		r.Protocol = getProtocolNameDto(r.Protocol)
		resultsSlice = append(resultsSlice, r)
	}
	return resultsSlice
}

func getProtocolNameDto(protocol string) string {
	switch protocol {
	case CCTP:
		return "cctp"
	case PortalTokenBridge:
		return "portal_token_bridge"
	case NTT:
		return strings.ToLower(NTT)
	default:
		return protocol
	}
}

func (s *Service) fetchProtocolValues(ctx context.Context, wg *sync.WaitGroup, protocol string, results chan<- ProtocolTotalValuesDTO, fetch func(context.Context, string) (ProtocolStats, error)) {
	defer wg.Done()

	val, err := cacheable.GetOrLoad[ProtocolStats](ctx,
		s.logger,
		s.cache,
		time.Duration(s.cacheTTL)*time.Minute,
		s.cacheKeyPrefix+":"+strings.ToUpper(protocol),
		s.metrics,
		func() (ProtocolStats, error) {
			return fetch(ctx, protocol)
		},
	)

	res := ProtocolTotalValuesDTO{
		ProtocolStats: val,
	}
	if err != nil {
		res.Error = err.Error()
	}
	results <- res
}

// getProtocolStats fetches stats for PortalTokenBridge and NTT
func (s *Service) getCoreProtocolStats(ctx context.Context, protocol string) (ProtocolStats, error) {

	protocolStats, err := s.getFetchFn(protocol)(ctx, protocol)
	if err != nil {
		return ProtocolStats{
			Protocol:              protocol,
			TotalValueTransferred: protocolStats.Latest.TotalValueTransferred,
			TotalMessages:         protocolStats.Latest.TotalMessages,
		}, err
	}

	diffLastDay := protocolStats.DeltaLast24hr.TotalMessages
	val := ProtocolStats{
		Protocol:              protocol,
		TotalValueTransferred: protocolStats.Latest.TotalValueTransferred,
		TotalMessages:         protocolStats.Latest.TotalMessages,
		LastDayMessages:       diffLastDay,
	}

	lastDayTotalMessages := protocolStats.Latest.TotalMessages - diffLastDay
	if lastDayTotalMessages != 0 {
		percentage := strconv.FormatFloat(float64(diffLastDay)/float64(lastDayTotalMessages)*100, 'f', 2, 64) + "%"
		val.LastDayDiffPercentage = percentage
	}

	if PortalTokenBridge == protocol {
		tvl, errTvl := s.tvl.Get(ctx)
		if errTvl != nil {
			s.logger.Error("error fetching tvl", zap.Error(errTvl), zap.String("protocol", protocol))
			return val, errTvl
		}
		tvlFloat, errTvl := strconv.ParseFloat(tvl, 64)
		if errTvl != nil {
			s.logger.Error("error parsing tvl value", zap.Error(errTvl), zap.String("protocol", protocol), zap.String("tvl_str", tvl))
			return val, errTvl
		}
		val.TotalValueLocked = tvlFloat
	}

	return val, nil
}

func (s *Service) getFetchFn(protocol string) fetchProtocolStats {
	switch protocol {
	case CCTP:
		return s.repo.getCCTPStats
	default:
		return s.repo.getCoreProtocolStats
	}
}

func (s *Service) getProtocolStats(ctx context.Context, protocol string) (ProtocolStats, error) {

	type statsResult struct {
		result stats
		Err    error
	}
	statsRes := make(chan statsResult, 1)
	go func() {
		defer close(statsRes)
		rowStats, errStats := s.repo.getProtocolStats(ctx, protocol)
		if errStats != nil {
			statsRes <- statsResult{Err: errStats}
			return
		}
		lastDayStats, errStats := s.repo.getProtocolStatsLastDay(ctx, protocol)
		if errStats != nil {
			statsRes <- statsResult{Err: errStats}
			return
		}
		statsRes <- statsResult{result: stats{Latest: rowStats, Last24: lastDayStats}}
	}()

	activity, err := s.repo.getProtocolActivity(ctx, protocol)
	if err != nil {
		s.logger.Error("error fetching protocol activity", zap.Error(err), zap.String("protocol", protocol))
		return ProtocolStats{Protocol: protocol}, err
	}

	rStats := <-statsRes
	if rStats.Err != nil {
		s.logger.Error("error fetching protocol stats", zap.Error(rStats.Err), zap.String("protocol", protocol))
		return ProtocolStats{Protocol: protocol}, rStats.Err
	}

	dto := ProtocolStats{
		Protocol:              protocol,
		TotalValueLocked:      rStats.result.Latest.TotalValueLocked,
		TotalMessages:         rStats.result.Latest.TotalMessages,
		TotalValueTransferred: activity.TotalValueTransferred,
		TotalValueSecured:     activity.TotalValueSecure,
	}

	totalMsgNow := rStats.result.Latest.TotalMessages
	totalMessagesAsFromLast24hr := rStats.result.Last24.TotalMessages
	if totalMessagesAsFromLast24hr != 0 {
		last24HrMessages := totalMsgNow - totalMessagesAsFromLast24hr
		dto.LastDayMessages = last24HrMessages
		dto.LastDayDiffPercentage = strconv.FormatFloat(float64(last24HrMessages)/float64(totalMessagesAsFromLast24hr)*100, 'f', 2, 64) + "%"
	}

	return dto, nil
}
