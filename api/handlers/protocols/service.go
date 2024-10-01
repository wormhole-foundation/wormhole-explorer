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
const MAYAN = "MAYAN"
const ALLBRIDGE = "ALLBRIDGE"

type Service struct {
	protocols      []string
	repo           *Repository
	logger         *zap.Logger
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
	Last24HourVolume      float64 `json:"last_24_hour_volume,omitempty"`
}

type tvlProvider interface {
	Get(ctx context.Context) (string, error)
}

type fetchProtocolStats func(ctx context.Context, protocol string) (intStats, error)
type fetchProtocolTotalValues func(context.Context, string) (ProtocolStats, error)

func NewService(protocols []string, repo *Repository, logger *zap.Logger, cache cache.Cache, cacheKeyPrefix string, cacheTTL int, metrics metrics.Metrics, tvlProvider tvlProvider) *Service {
	return &Service{
		protocols:      protocols,
		repo:           repo,
		logger:         logger,
		cache:          cache,
		cacheKeyPrefix: cacheKeyPrefix,
		cacheTTL:       cacheTTL,
		metrics:        metrics,
		tvl:            tvlProvider,
	}
}

func (s *Service) GetProtocolsTotalValues(ctx context.Context) []ProtocolTotalValuesDTO {

	protocolsQty := len(s.protocols)
	results := make(chan ProtocolTotalValuesDTO, protocolsQty)

	wg := &sync.WaitGroup{}
	wg.Add(protocolsQty)

	for _, p := range s.protocols {
		fetchFn := s.getProtocolTotalValuesFn(p)
		go s.fetchProtocolValues(ctx, wg, p, results, fetchFn)
	}

	wg.Wait()
	close(results)

	resultsSlice := make([]ProtocolTotalValuesDTO, 0, len(s.protocols))
	for r := range results {
		r.Protocol = getProtocolNameDto(r.Protocol)
		resultsSlice = append(resultsSlice, r)
	}
	return resultsSlice
}

func (s *Service) fetchProtocolValues(ctx context.Context, wg *sync.WaitGroup, protocol string, results chan<- ProtocolTotalValuesDTO, fetch fetchProtocolTotalValues) {
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

func (s *Service) getProtocolTotalValuesFn(protocol string) fetchProtocolTotalValues {
	switch protocol {
	case MAYAN:
		return s.getMayanStats
	case ALLBRIDGE:
		return s.getAllbridgeStats
	default:
		return s.getCoreProtocolStats
	}
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

// getProtocolStats fetches stats for PortalTokenBridge and NTT
func (s *Service) getCoreProtocolStats(ctx context.Context, protocol string) (ProtocolStats, error) {

	protocolStats, err := s.getRepositoryFetchFn(protocol)(ctx, protocol)
	if err != nil {
		return ProtocolStats{
			Protocol: protocol,
		}, err
	}

	diffLastDay := protocolStats.DeltaLast24hr.TotalMessages
	val := ProtocolStats{
		Protocol:              protocol,
		TotalValueTransferred: protocolStats.Latest.TotalValueTransferred,
		TotalMessages:         protocolStats.Latest.TotalMessages,
		LastDayMessages:       diffLastDay,
		Last24HourVolume:      protocolStats.DeltaLast24hr.TotalValueTransferred,
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

func (s *Service) getRepositoryFetchFn(protocol string) fetchProtocolStats {
	switch protocol {
	case CCTP:
		return s.repo.getCCTPStats
	default:
		return s.repo.getCoreProtocolStats
	}
}

func (s *Service) getMayanStats(ctx context.Context, _ string) (ProtocolStats, error) {
	const mayan = "mayan"
	mayanNow, errStats := s.repo.getProtocolStatsNow(ctx, mayan)
	if errStats != nil {
		s.logger.Error("error fetching Mayan stats", zap.Error(errStats), zap.String("protocol", mayan))
		return ProtocolStats{Protocol: mayan}, errStats
	}
	mayan24hrAgo, errStats := s.repo.getProtocolStats24hrAgo(ctx, mayan)
	if errStats != nil {
		s.logger.Error("error fetching Mayan stats 24hr ago", zap.Error(errStats), zap.String("protocol", mayan))
		return ProtocolStats{Protocol: mayan}, errStats
	}

	last24HrMessages := mayanNow.TotalMessages - mayan24hrAgo.TotalMessages
	return ProtocolStats{
		Protocol:              mayan,
		TotalValueLocked:      mayanNow.TotalValueLocked,
		TotalMessages:         mayanNow.TotalMessages,
		TotalValueTransferred: mayanNow.Volume,
		LastDayMessages:       last24HrMessages,
		Last24HourVolume:      mayanNow.Volume - mayan24hrAgo.Volume,
		LastDayDiffPercentage: strconv.FormatFloat(float64(last24HrMessages)/float64(mayan24hrAgo.TotalMessages)*100, 'f', 2, 64) + "%",
	}, nil
}

func (s *Service) getAllbridgeStats(ctx context.Context, _ string) (ProtocolStats, error) {

	const allbridge = "allbridge"
	type statsResult struct {
		result stats
		Err    error
	}
	statsRes := make(chan statsResult, 1)
	go func() {
		defer close(statsRes)
		statsNow, errStats := s.repo.getProtocolStatsNow(ctx, allbridge)
		if errStats != nil {
			statsRes <- statsResult{Err: errStats}
			return
		}
		stats24hrAgo, errStats := s.repo.getProtocolStats24hrAgo(ctx, allbridge)
		if errStats != nil {
			statsRes <- statsResult{Err: errStats}
			return
		}
		statsRes <- statsResult{result: stats{Latest: statsNow, Last24: stats24hrAgo}}
	}()

	activity, err := s.repo.getAllbridgeActivity(ctx)
	if err != nil {
		s.logger.Error("error fetching allbridge activity", zap.Error(err), zap.String("protocol", allbridge))
		return ProtocolStats{Protocol: allbridge}, err
	}

	rStats := <-statsRes
	if rStats.Err != nil {
		s.logger.Error("error fetching allbridge stats", zap.Error(rStats.Err), zap.String("protocol", allbridge))
		return ProtocolStats{Protocol: allbridge}, rStats.Err
	}

	dto := ProtocolStats{
		Protocol:              allbridge,
		TotalMessages:         rStats.result.Latest.TotalMessages,
		TotalValueTransferred: activity.TotalValueTransferred,
		Last24HourVolume:      activity.Last24HrTotalValueTransferred,
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
