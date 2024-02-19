package stats

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"time"
)

type ProtocolsStatsJob struct {
	statsDB              api.WriteAPIBlocking
	logger               *zap.Logger
	statsClientsFetchers []ClientStats
	version              string
}

type Stats struct {
	TotalValueLocked string `json:"total_value_locked"`
	TotalMessages    string `json:"total_messages"`
}

// ClientStats Abstraction for fetching stats since each protocol may have different implementation details.
type ClientStats interface {
	Get(ctx context.Context) (Stats, error)
	ProtocolName() string
}

type protocolStats struct {
	Stats
	Name string
}

// NewProtocolsStatsJob creates an instance of the job implementation.
func NewProtocolsStatsJob(statsDB api.WriteAPIBlocking, logger *zap.Logger, version string, statsFetchers ...ClientStats) *ProtocolsStatsJob {
	return &ProtocolsStatsJob{
		statsDB:              statsDB,
		logger:               logger.With(zap.String("module", "ProtocolsStatsJob")),
		statsClientsFetchers: statsFetchers,
		version:              version,
	}
}

func (s *ProtocolsStatsJob) Run(ctx context.Context) error {

	clientsQty := len(s.statsClientsFetchers)
	wg := sync.WaitGroup{}
	wg.Add(clientsQty)
	stats := make(chan protocolStats, clientsQty)
	var anyError error

	for _, cs := range s.statsClientsFetchers {
		go func(c ClientStats) {
			defer wg.Done()
			st, err := c.Get(ctx)
			if err != nil {
				anyError = err
				return
			}
			stats <- protocolStats{st, c.ProtocolName()}
		}(cs)
	}

	wg.Wait()
	close(stats)

	err := s.updateStats(ctx, stats)
	if err != nil {
		anyError = err
	}

	return anyError
}

func (s *ProtocolsStatsJob) updateStats(ctx context.Context, stats <-chan protocolStats) error {

	ts := time.Now().UTC().Truncate(time.Hour) // make minutes and seconds zero, so we only work with date and hour
	points := make([]*write.Point, 0, len(stats))

	for st := range stats {
		point := influxdb2.
			NewPointWithMeasurement(dbconsts.ProtocolsStatsMeasurement).
			AddTag("protocol", st.Name).
			AddTag("version", s.version).
			AddField("total_messages", st.TotalMessages).
			AddField("total_value_locked", st.TotalValueLocked).
			SetTime(ts)

		points = append(points, point)
	}

	err := s.statsDB.WritePoint(ctx, points...)
	if err != nil {
		s.logger.Error("failed updating protocol stats in influxdb", zap.Error(err))
	}
	return err
}

// Default implementation of ClientStats interface. Encapsulate the url and http.client for calling a specific external service to retrieve stats
type httpRestClientStats struct {
	name   string
	url    string
	client httpDo
	logger *zap.Logger
}

type httpDo interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHttpRestClientStats(name, url string, logger *zap.Logger, httpClient httpDo) ClientStats {
	return &httpRestClientStats{
		name:   name,
		url:    url,
		logger: logger,
		client: httpClient,
	}
}

func (d *httpRestClientStats) ProtocolName() string {
	return d.name
}

func (d *httpRestClientStats) Get(ctx context.Context) (Stats, error) {

	decoratedLogger := d.logger

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving client stats",
			zap.Error(err),
		)
		return Stats{}, errors.WithStack(err)
	}

	reqId := uuid.New().String()
	req.Header.Set("X-Request-ID", reqId)
	decoratedLogger = decoratedLogger.With(zap.String("requestID", reqId))

	resp, err := d.client.Do(req)
	if err != nil {
		decoratedLogger.Error("failed retrieving client stats",
			zap.Error(err),
		)
		return Stats{}, errors.WithStack(err)
	}
	defer resp.Body.Close()

	decoratedLogger = decoratedLogger.
		With(zap.String("status_code", http.StatusText(resp.StatusCode))).
		With(zap.String("response_headers", toJson(resp.Header)))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		decoratedLogger.Error("error retrieving client stats: got an invalid response status code",
			zap.String("response_body", string(body)),
		)
		return Stats{}, errors.Errorf("failed retrieving client stats from url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from client stats. url:%s - status_code:%d", d.url, resp.StatusCode)
	}
	var stats Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from client stats. url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}
	return stats, nil

}

func toJson(headers http.Header) string {
	bytes, _ := json.Marshal(headers)
	return string(bytes)
}
