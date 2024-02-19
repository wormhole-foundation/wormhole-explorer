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

type ProtocolsActivityJob struct {
	statsDB          api.WriteAPIBlocking
	logger           *zap.Logger
	activityFetchers []ClientActivity
	version          string
}

type ProtocolActivity struct {
	TotalValueSecured     string `json:"total_value_secured"`
	TotalValueTransferred string `json:"total_value_transferred"`
	Activity              []struct {
		EmitterChainID     string `json:"emitter_chain_id"`
		DestinationChainID string `json:"destination_chain_id"`
		Txs                string `json:"txs"`
		TotalUSD           string `json:"total_usd"`
	} `json:"activity"`
}

// ClientActivity Abstraction for fetching protocol activity since each client may have different implementation details.
type ClientActivity interface {
	Get(ctx context.Context, from, to time.Time) (ProtocolActivity, error)
	ProtocolName() string
}

// NewProtocolActivityJob creates an instance of the job implementation.
func NewProtocolActivityJob(statsDB api.WriteAPIBlocking, logger *zap.Logger, version string, activityFetchers ...ClientActivity) *ProtocolsActivityJob {
	return &ProtocolsActivityJob{
		statsDB:          statsDB,
		logger:           logger.With(zap.String("module", "ProtocolsActivityJob")),
		activityFetchers: activityFetchers,
		version:          version,
	}
}

func (m *ProtocolsActivityJob) Run(ctx context.Context) error {

	clientsQty := len(m.activityFetchers)
	wg := sync.WaitGroup{}
	wg.Add(clientsQty)
	errs := make(chan error, clientsQty)
	ts := time.Now().UTC().Truncate(time.Hour) // make minutes and seconds zero, so we only work with date and hour
	from := ts.Add(-1 * time.Hour)

	for _, cs := range m.activityFetchers {
		go func(c ClientActivity) {
			defer wg.Done()
			activity, err := c.Get(ctx, from, ts)
			if err != nil {
				errs <- err
				return
			}
			errs <- m.updateActivity(ctx, c.ProtocolName(), m.version, activity, from)
		}(cs)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ProtocolsActivityJob) updateActivity(ctx context.Context, protocol, version string, activity ProtocolActivity, ts time.Time) error {

	points := make([]*write.Point, 0, len(activity.Activity))

	for i := range activity.Activity {
		point := influxdb2.
			NewPointWithMeasurement(dbconsts.ProtocolsActivityMeasurement).
			AddTag("protocol", protocol).
			AddTag("emitter_chain_id", activity.Activity[i].EmitterChainID).
			AddTag("destination_chain_id", activity.Activity[i].DestinationChainID).
			AddTag("version", version).
			AddField("total_volume_secure", activity.TotalValueSecured).
			AddField("total_value_transferred", activity.TotalValueTransferred).
			AddField("txs", activity.Activity[i].Txs).
			AddField("total_usd", activity.Activity[i].TotalUSD).
			SetTime(ts)
		points = append(points, point)
	}

	err := m.statsDB.WritePoint(ctx, points...)
	if err != nil {
		m.logger.Error("failed updating protocol activity in influxdb", zap.Error(err), zap.String("protocol", protocol))
	}
	return err
}

// Default implementation of ClientActivity interface. Encapsulate the url and http.client for calling a specific external service to retrieve activity
type httpRestClientActivity struct {
	name   string
	url    string
	client httpDo
	logger *zap.Logger
}

func NewHttpRestClientActivity(name, url string, logger *zap.Logger, httpClient httpDo) ClientActivity {
	return &httpRestClientActivity{
		name:   name,
		url:    url,
		logger: logger,
		client: httpClient,
	}
}

func (d *httpRestClientActivity) ProtocolName() string {
	return d.name
}

func (d *httpRestClientActivity) Get(ctx context.Context, from, to time.Time) (ProtocolActivity, error) {

	decoratedLogger := d.logger

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving protocol activity",
			zap.Error(err),
		)
		return ProtocolActivity{}, errors.WithStack(err)
	}
	q := req.URL.Query()
	q.Set("from", from.Format(time.RFC3339))
	q.Set("to", to.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	reqId := uuid.New().String()
	req.Header.Set("X-Request-ID", reqId)
	decoratedLogger = decoratedLogger.With(zap.String("requestID", reqId))

	resp, err := d.client.Do(req)
	if err != nil {
		decoratedLogger.Error("failed retrieving protocol activity",
			zap.Error(err),
		)
		return ProtocolActivity{}, errors.WithStack(err)
	}
	defer resp.Body.Close()

	decoratedLogger = decoratedLogger.
		With(zap.String("status_code", http.StatusText(resp.StatusCode))).
		With(zap.String("response_headers", toJson(resp.Header)))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		decoratedLogger.Error("error retrieving protocol activity: got an invalid response status code",
			zap.String("response_body", string(body)),
		)
		return ProtocolActivity{}, errors.Errorf("failed retrieving protocol activity from url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol activity. url:%s - status_code:%d", d.url, resp.StatusCode)
	}
	var protocolActivity ProtocolActivity
	err = json.Unmarshal(body, &protocolActivity)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol activity. url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}
	return protocolActivity, nil
}
