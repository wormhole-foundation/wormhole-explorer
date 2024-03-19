package repository

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons"
	"go.uber.org/zap"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

func NewAllBridgeRestClient(baseURL string, logger *zap.Logger, httpClient commons.HttpDo) *AllBridgeRestClient {
	return &AllBridgeRestClient{
		baseURL: baseURL,
		logger:  logger,
		client:  httpClient,
	}
}

type AllBridgeRestClient struct {
	baseURL string
	client  commons.HttpDo
	logger  *zap.Logger
}

type allBridgeActivity struct {
	TotalValueSecured     string `json:"total_value_secure"`
	TotalValueTransferred string `json:"total_value_transferred"`
	Activities            []struct {
		EmitterChainID     uint64 `json:"emitter_chain_id"`
		DestinationChainID uint64 `json:"destination_chain_id"`
		Txs                string `json:"txs"`
		TotalUSD           string `json:"total_usd"`
	} `json:"activity"`
}

func (d *AllBridgeRestClient) ProtocolName() string {
	return commons.AllBridgeProtocol
}

func (d *AllBridgeRestClient) GetActivity(ctx context.Context, from, to time.Time) (ProtocolActivity, error) {
	decoratedLogger := d.logger

	url := d.baseURL + "/wormhole/activity"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving protocol Activities",
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
		decoratedLogger.Error("failed retrieving protocol Activities",
			zap.Error(err),
		)
		return ProtocolActivity{}, errors.WithStack(err)
	}
	defer resp.Body.Close()

	decoratedLogger = decoratedLogger.
		With(zap.String("status_code", http.StatusText(resp.StatusCode))).
		With(zap.String("response_headers", commons.ToJson(resp.Header)))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		decoratedLogger.Error("error retrieving protocol Activities: got an invalid response status code",
			zap.String("response_body", string(body)), zap.Int("status_code", resp.StatusCode),
		)
		return ProtocolActivity{}, errors.Errorf("failed retrieving protocol Activities from baseURL:%s - status_code:%d - response_body:%s", url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol Activities. baseURL:%s - status_code:%d", d.baseURL, resp.StatusCode)
	}

	var temp allBridgeActivity
	err = json.Unmarshal(body, &temp)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol Activities. baseURL:%s - status_code:%d - response_body:%s", d.baseURL, resp.StatusCode, string(body))
	}

	return temp.toProtocolActivity()
}

func (d *AllBridgeRestClient) GetStats(ctx context.Context) (Stats, error) {

	decoratedLogger := d.logger

	url := d.baseURL + "/wormhole/stats"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving protocol stats", zap.Error(err))
		return Stats{}, errors.WithStack(err)
	}

	reqId := uuid.New().String()
	req.Header.Set("X-Request-ID", reqId)
	decoratedLogger = decoratedLogger.With(zap.String("requestID", reqId))

	resp, err := d.client.Do(req)
	if err != nil {
		decoratedLogger.Error("failed retrieving protocol stats", zap.Error(err))
		return Stats{}, errors.WithStack(err)
	}
	defer resp.Body.Close()

	decoratedLogger = decoratedLogger.
		With(zap.String("status_code", http.StatusText(resp.StatusCode))).
		With(zap.String("response_headers", commons.ToJson(resp.Header)))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		decoratedLogger.Error("error retrieving protocol stats: got an invalid response status code", zap.String("response_body", string(body)), zap.Int("status_code", resp.StatusCode))
		return Stats{}, errors.Errorf("failed retrieving protocol stats from baseURL:%s - status_code:%d - response_body:%s", d.baseURL, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol stats. baseURL:%s - status_code:%d", d.baseURL, resp.StatusCode)
	}

	var allbridgeStats allBridgeStatsResponseDTO
	err = json.Unmarshal(body, &allbridgeStats)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol stats. baseURL:%s - status_code:%d - response_body:%s", d.baseURL, resp.StatusCode, string(body))
	}

	return d.toStats(allbridgeStats)
}

func (m *allBridgeActivity) toProtocolActivity() (ProtocolActivity, error) {
	result := ProtocolActivity{}

	totalValueSecured, err := strconv.ParseFloat(m.TotalValueSecured, 64)
	if err != nil || math.IsNaN(totalValueSecured) {
		return result, errors.Wrap(err, "failed parsing string TotalValueSecure to float64")
	}
	result.TotalValueSecure = totalValueSecured

	totalValueTransferred, err := strconv.ParseFloat(m.TotalValueTransferred, 64)
	if err != nil || math.IsNaN(totalValueTransferred) {
		return result, errors.Wrap(err, "failed parsing string TotalValueTransferred to float64")
	}
	result.TotalValueTransferred = totalValueTransferred

	for i := range m.Activities {

		act := m.Activities[i]
		txs, errTxs := strconv.ParseUint(act.Txs, 10, 64)
		if errTxs != nil {
			return result, errors.Wrap(errTxs, "failed parsing string txs to uint64")
		}

		totalUSD, errTotalUSD := strconv.ParseFloat(act.TotalUSD, 64)
		if errTotalUSD != nil {
			return result, errors.Wrap(errTxs, "failed parsing string total_usd to float64")
		}

		a := Activity{
			EmitterChainID:     m.Activities[i].EmitterChainID,
			DestinationChainID: m.Activities[i].DestinationChainID,
			Txs:                txs,
			TotalUSD:           totalUSD,
		}
		result.Activities = append(result.Activities, a)
	}

	return result, nil
}

type allBridgeStatsResponseDTO struct {
	TotalValueLocked string `json:"total_value_locked"`
	TotalMessages    string `json:"total_messages"`
	Volume           string `json:"volume"`
}

func (d *AllBridgeRestClient) toStats(t allBridgeStatsResponseDTO) (Stats, error) {

	convertAndLoad := func(val string, target *float64) error {
		if len(val) == 0 {
			*target = 0
			return nil
		}
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			d.logger.Error("failed converting value", zap.Error(err), zap.String("value", val))
			return err
		}
		*target = floatVal
		return nil
	}

	var stats Stats

	err := convertAndLoad(t.TotalValueLocked, &stats.TotalValueLocked)
	if err != nil {
		return stats, err
	}

	err = convertAndLoad(t.Volume, &stats.Volume)
	if err != nil {
		return stats, err
	}

	totalMsg, err := strconv.ParseUint(t.TotalMessages, 10, 64)
	if err != nil {
		return stats, err
	}
	stats.TotalMessages = totalMsg

	return stats, nil

}
