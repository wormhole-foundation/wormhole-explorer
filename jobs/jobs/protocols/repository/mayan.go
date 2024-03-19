package repository

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

func NewMayanRestClient(baseURL string, logger *zap.Logger, httpClient commons.HttpDo) *MayanRestClient {
	return &MayanRestClient{
		baseURL: baseURL,
		logger:  logger,
		client:  httpClient,
	}
}

type MayanRestClient struct {
	baseURL string
	client  commons.HttpDo
	logger  *zap.Logger
}

func (d *MayanRestClient) GetActivity(ctx context.Context, from, to time.Time) (ProtocolActivity, error) {
	decoratedLogger := d.logger

	url := d.baseURL + "/v3/stats/wh/activity"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving protocol activities",
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
		decoratedLogger.Error("failed retrieving protocol activities",
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
		decoratedLogger.Error("error retrieving protocol activities: got an invalid response status code",
			zap.String("response_body", string(body)),
		)
		return ProtocolActivity{}, errors.Errorf("failed retrieving protocol activities from url:%s - status_code:%d - response_body:%s", url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol activities. url:%s - status_code:%d", url, resp.StatusCode)
	}

	type mayanActivity struct {
		ProtocolActivity
		Activities []struct {
			AlternativeEmitterChainID string  `json:"emmiter_chain_id"` // typo is on purpose due to mayan-api returning in that format
			DestinationChainID        string  `json:"destination_chain_id"`
			Txs                       uint64  `json:"txs"`
			TotalUSD                  float64 `json:"total_usd"`
		} `json:"activity"`
	}

	var mayanResp mayanActivity
	err = json.Unmarshal(body, &mayanResp)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return ProtocolActivity{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol activities. url:%s - status_code:%d - response_body:%s", url, resp.StatusCode, string(body))
	}

	result := ProtocolActivity{
		TotalValueTransferred: mayanResp.TotalValueTransferred,
		TotalValueSecure:      mayanResp.TotalValueSecure,
		TotalMessages:         mayanResp.TotalMessages,
		Volume:                mayanResp.Volume,
	}

	for _, act := range mayanResp.Activities {

		emitterChainId, errEmitter := strconv.ParseUint(act.AlternativeEmitterChainID, 10, 64)
		if errEmitter != nil {
			return ProtocolActivity{}, errors.Wrap(errEmitter, "failed parsing protocol activity emitter chain id from string to uint64")
		}

		destChainId, errDest := strconv.ParseUint(act.DestinationChainID, 10, 64)
		if errDest != nil {
			return ProtocolActivity{}, errors.Wrap(errDest, "failed parsing protocol activity destination chain id from string to uint64")
		}

		val := Activity{
			EmitterChainID:     emitterChainId,
			DestinationChainID: destChainId,
			Txs:                act.Txs,
			TotalUSD:           act.TotalUSD,
		}
		result.Activities = append(result.Activities, val)
	}

	return result, nil
}

func (d *MayanRestClient) GetStats(ctx context.Context) (Stats, error) {
	decoratedLogger := d.logger
	url := d.baseURL + "/v3/stats/wh/stats"
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
		decoratedLogger.Error("error retrieving client stats: got an invalid response status code", zap.String("response_body", string(body)))
		return Stats{}, errors.Errorf("failed retrieving protocol stats from url:%s - status_code:%d - response_body:%s", url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol stats. url:%s - status_code:%d", url, resp.StatusCode)
	}
	var stats Stats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return Stats{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol stats. url:%s - status_code:%d - response_body:%s", url, resp.StatusCode, string(body))
	}

	return stats, nil
}

func (d *MayanRestClient) ProtocolName() string {
	return commons.MayanProtocol
}
