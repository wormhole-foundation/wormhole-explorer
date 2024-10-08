package protocols

import (
	"context"
	"fmt"
	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CCTP = "CCTP_WORMHOLE_INTEGRATION"
const PortalTokenBridge = "PORTAL_TOKEN_BRIDGE"
const GenericRelayer = "GENERIC_RELAYER"
const NTT = "NATIVE_TOKEN_TRANSFER"
const MAYAN = "MAYAN"
const ALLBRIDGE = "ALLBRIDGE"
const UNKNOWN = "UNKNOWN"

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
	Protocol                    string  `json:"protocol"`
	TotalMessages               uint64  `json:"total_messages"`
	TotalValueLocked            float64 `json:"total_value_locked,omitempty"`
	TotalValueSecured           float64 `json:"total_value_secured,omitempty"`
	TotalValueTransferred       float64 `json:"total_value_transferred"`
	LastDayMessages             uint64  `json:"last_day_messages"`
	LastDayDiffPercentage       string  `json:"last_day_diff_percentage"`
	LastDayVolumeDiffPercentage string  `json:"last_day_diff_volume_percentage"`
	Last24HourVolume            float64 `json:"last_24_hour_volume"`
}

type tvlProvider interface {
	Get(ctx context.Context) (string, error)
}

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
	results := make(chan ProtocolTotalValuesDTO) // unbuffered

	wg := &sync.WaitGroup{}
	wg.Add(protocolsQty)

	for _, p := range s.protocols {
		fetchFn := s.getProtocolTotalValuesFn(p)
		go s.fetchProtocolValues(ctx, wg, p, results, fetchFn) // fetch protocols which are populated from other sources: CCTP, MAYAN, ALLBRIDGE
	}

	wg.Add(1)
	go s.fetchAllProtocolValues(ctx, wg, s.protocols, results) // fetch all protocols from the vaas we received from gossip network

	go func() {
		wg.Wait() // wait for all goroutines to finish in order to close the channel
		close(results)
	}()

	resultsSlice := make([]ProtocolTotalValuesDTO, 0, protocolsQty)
	for r := range results {
		if r.Protocol == UNKNOWN {
			continue
		}
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
		cacheable.WithAutomaticRenew(),
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
	case CCTP:
		return s.getCCTPStats
	default:
		return func(_ context.Context, _ string) (ProtocolStats, error) {
			return ProtocolStats{Protocol: protocol}, fmt.Errorf("unsupported protocol %s", protocol)
		}
	}
}

func (s *Service) fetchAllProtocolValues(ctx context.Context, wg *sync.WaitGroup, excludedProtocols []string, results chan<- ProtocolTotalValuesDTO) {
	defer wg.Done()

	val, err := cacheable.GetOrLoad[[]ProtocolStats](ctx,
		s.logger,
		s.cache,
		time.Duration(s.cacheTTL)*time.Minute,
		s.cacheKeyPrefix+":ALL_PROTOCOLS",
		s.metrics,
		func() ([]ProtocolStats, error) {
			return s.getAllProtocolStats(ctx, excludedProtocols)
		},
		cacheable.WithAutomaticRenew(),
	)

	if err != nil {
		results <- ProtocolTotalValuesDTO{Error: err.Error()}
		return
	}

	for _, v := range val {
		results <- ProtocolTotalValuesDTO{ProtocolStats: v}
	}
}

func (s *Service) getAllProtocolStats(ctx context.Context, excludeProtocols []string) ([]ProtocolStats, error) {

	allProtocolsStats, err := s.repo.getAllProtocolStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ProtocolStats, 0, len(allProtocolsStats))
	for _, protocolStats := range allProtocolsStats {

		protocol := protocolStats.Latest.Protocol

		if slices.Contains(excludeProtocols, protocol) {
			continue
		}

		diffLastDay := protocolStats.DeltaLast24hr.TotalMessages
		diffVolumeLastDay := protocolStats.DeltaLast24hr.TotalValueTransferred

		val := ProtocolStats{
			Protocol:                    protocol,
			TotalValueTransferred:       protocolStats.Latest.TotalValueTransferred,
			TotalMessages:               protocolStats.Latest.TotalMessages,
			LastDayMessages:             diffLastDay,
			Last24HourVolume:            protocolStats.DeltaLast24hr.TotalValueTransferred,
			LastDayDiffPercentage:       "0.00%",
			LastDayVolumeDiffPercentage: "0.00%",
		}

		lastDayTotalMessages := protocolStats.Latest.TotalMessages - diffLastDay
		if lastDayTotalMessages != 0 {
			percentage := strconv.FormatFloat(float64(diffLastDay)/float64(lastDayTotalMessages)*100, 'f', 2, 64) + "%"
			val.LastDayDiffPercentage = percentage
		}

		volume24HourAgo := protocolStats.Latest.TotalValueTransferred - diffVolumeLastDay
		if volume24HourAgo != 0 {
			percentage := strconv.FormatFloat(diffVolumeLastDay/volume24HourAgo*100, 'f', 2, 64) + "%"
			val.LastDayVolumeDiffPercentage = percentage
		}

		if PortalTokenBridge == protocol {
			tvl, errTvl := s.tvl.Get(ctx)
			if errTvl != nil {
				s.logger.Error("error fetching tvl", zap.Error(errTvl), zap.String("protocol", protocol))
				return result, errTvl
			}
			tvlFloat, errTvl := strconv.ParseFloat(tvl, 64)
			if errTvl != nil {
				s.logger.Error("error parsing tvl value", zap.Error(errTvl), zap.String("protocol", protocol), zap.String("tvl_str", tvl))
				return result, errTvl
			}
			val.TotalValueLocked = tvlFloat
		}
		result = append(result, val)
	}
	return result, nil
}

func getProtocolNameDto(protocol string) string {
	switch protocol {
	case CCTP:
		return "cctp"
	case PortalTokenBridge:
		return "portal_token_bridge"
	case GenericRelayer:
		return "standard_relayer"
	default:
		return strings.ToLower(protocol)
	}
}

func (s *Service) getCCTPStats(ctx context.Context, protocol string) (ProtocolStats, error) {
	cctpStats, err := s.repo.getCCTPStats(ctx, protocol)
	if err != nil {
		return ProtocolStats{
			Protocol: protocol,
		}, err
	}

	diffLastDay := cctpStats.DeltaLast24hr.TotalMessages
	val := ProtocolStats{
		Protocol:                    protocol,
		TotalValueTransferred:       cctpStats.Latest.TotalValueTransferred,
		TotalMessages:               cctpStats.Latest.TotalMessages,
		LastDayMessages:             diffLastDay,
		Last24HourVolume:            cctpStats.DeltaLast24hr.TotalValueTransferred,
		LastDayDiffPercentage:       "0.00%",
		LastDayVolumeDiffPercentage: "0.00%",
	}

	lastDayTotalMessages := cctpStats.Latest.TotalMessages - diffLastDay
	if lastDayTotalMessages != 0 {
		percentage := strconv.FormatFloat(float64(diffLastDay)/float64(lastDayTotalMessages)*100, 'f', 2, 64) + "%"
		val.LastDayDiffPercentage = percentage
	}

	lastDayVolume := cctpStats.Latest.TotalValueTransferred - cctpStats.DeltaLast24hr.TotalValueTransferred
	if lastDayVolume != 0 {
		percentage := strconv.FormatFloat(cctpStats.DeltaLast24hr.TotalValueTransferred/lastDayVolume*100, 'f', 2, 64) + "%"
		val.LastDayVolumeDiffPercentage = percentage
	}

	return val, nil
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
	last24HrVolume := mayanNow.Volume - mayan24hrAgo.Volume
	return ProtocolStats{
		Protocol:                    mayan,
		TotalValueLocked:            mayanNow.TotalValueLocked,
		TotalMessages:               mayanNow.TotalMessages,
		TotalValueTransferred:       mayanNow.Volume,
		LastDayMessages:             last24HrMessages,
		Last24HourVolume:            mayanNow.Volume - mayan24hrAgo.Volume,
		LastDayDiffPercentage:       strconv.FormatFloat(float64(last24HrMessages)/float64(mayan24hrAgo.TotalMessages)*100, 'f', 2, 64) + "%",
		LastDayVolumeDiffPercentage: strconv.FormatFloat(last24HrVolume/mayan24hrAgo.Volume*100, 'f', 2, 64) + "%",
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
		Protocol:                    allbridge,
		TotalMessages:               rStats.result.Latest.TotalMessages,
		TotalValueTransferred:       activity.TotalValueTransferred,
		Last24HourVolume:            activity.Last24HrTotalValueTransferred,
		LastDayDiffPercentage:       "0.00%",
		LastDayVolumeDiffPercentage: "0.00%",
	}

	totalMsgNow := rStats.result.Latest.TotalMessages
	totalMessagesAsFromLast24hr := rStats.result.Last24.TotalMessages
	if totalMessagesAsFromLast24hr != 0 {
		last24HrMessages := totalMsgNow - totalMessagesAsFromLast24hr
		dto.LastDayMessages = last24HrMessages
		dto.LastDayDiffPercentage = strconv.FormatFloat(float64(last24HrMessages)/float64(totalMessagesAsFromLast24hr)*100, 'f', 2, 64) + "%"
	}

	totalVolumeAsFromLast24Hr := activity.TotalValueTransferred - activity.Last24HrTotalValueTransferred
	if totalVolumeAsFromLast24Hr != 0 {
		last24HrVolume := activity.Last24HrTotalValueTransferred
		dto.LastDayVolumeDiffPercentage = strconv.FormatFloat(last24HrVolume/totalVolumeAsFromLast24Hr*100, 'f', 2, 64) + "%"
	}

	return dto, nil
}
