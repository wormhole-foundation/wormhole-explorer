package stats

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
)

const contributorsStatsMeasurement = "contributors_stats"

type ContributorsStatsJob struct {
	statsDB              api.WriteAPIBlocking
	logger               *zap.Logger
	statsClientsFetchers []ClientStats
}

type ClientStatsInfo struct {
	Url  string
	Name string
}

type Stats struct {
	TotalValueLocked string `json:"total_value_locked"`
	TotalMessages    string `json:"total_messages"`
}

// ClientStats Abstraction for fetching stats since each client may have different implementation details.
type ClientStats interface {
	Get(ctx context.Context) (Stats, error)
	ContributorName() string
}

// NewContributorsStatsJob creates an instance of the job implementation.
// func NewContributorsStatsJob(statsDB api.WriteAPIBlocking, logger *zap.Logger, contributorsInfo ...ClientStatsInfo) *ContributorsStatsJob {
func NewContributorsStatsJob(statsDB api.WriteAPIBlocking, logger *zap.Logger, statsFetchers ...ClientStats) *ContributorsStatsJob {
	return &ContributorsStatsJob{
		statsDB: statsDB,
		logger:  logger.With(zap.String("module", "ContributorsStatsJob")),
		//statsClientsFetchers: createStatsFetchers(logger, contributorsInfo),
		statsClientsFetchers: statsFetchers,
	}
}

func (s *ContributorsStatsJob) Run(ctx context.Context) error {

	clientsQty := len(s.statsClientsFetchers)
	wg := sync.WaitGroup{}
	wg.Add(clientsQty)
	errs := make(chan error, clientsQty)

	for _, cs := range s.statsClientsFetchers {
		go func(c ClientStats) {
			defer wg.Done()
			stats, err := c.Get(ctx)
			if err != nil {
				errs <- err
				return
			}
			errs <- s.updateStats(ctx, c.ContributorName(), stats)
		}(cs)
	}

	wg.Wait()
	close(errs)

	var err error
	for e := range errs {
		if e != nil {
			err = e
		}
	}

	return err
}

func (s *ContributorsStatsJob) updateStats(ctx context.Context, serviceName string, stats Stats) error {

	point := influxdb2.
		NewPointWithMeasurement(contributorsStatsMeasurement).
		AddTag("contributor", serviceName).
		AddField("total_messages", stats.TotalMessages).
		AddField("total_value_locked", stats.TotalValueLocked)

	err := s.statsDB.WritePoint(ctx, point)
	if err != nil {
		s.logger.Error("failed updating contributor stats in influxdb", zap.Error(err))
	}
	return err
}

/*
func createStatsFetchers(logger *zap.Logger, infos []ClientStatsInfo) []ClientStats {
	fetchers := make([]ClientStats, 0, len(infos))
	for _, cInfo := range infos {
		fetchers = append(fetchers, &httpRestClientStats{
			url:        cInfo.Url,
			client:     &http.Client{},
			logger:     logger.With(zap.String("sevice", cInfo.Name), zap.String("url", cInfo.Url)),
			name: cInfo.Name})
	}
	return fetchers
}
*/

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

func (d *httpRestClientStats) ContributorName() string {
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

	respBody, _ := io.ReadAll(resp.Body) // skip handling the error since the client may not send a response body

	if resp.StatusCode != http.StatusOK {

		decoratedLogger.Error("error retrieving client stats: got an invalid response status code",
			zap.String("response_body", string(respBody)),
		)
		return Stats{}, errors.Errorf("failed retrieving client stats from url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(respBody))
	}

	var stats Stats
	err = json.Unmarshal(respBody, &stats)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from client stats. url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(respBody))
	}
	return stats, nil

}

func toJson(headers http.Header) string {
	bytes, _ := json.Marshal(headers)
	return string(bytes)
}
