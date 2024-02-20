package repositories

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

func NewAllBridgeRestClient(name, url string, logger *zap.Logger, httpClient commons.HttpDo) *AllBridgeRestClient {
	return &AllBridgeRestClient{
		name:   name,
		url:    url,
		logger: logger,
		client: httpClient,
	}
}

type AllBridgeRestClient struct {
	name   string
	url    string
	client commons.HttpDo
	logger *zap.Logger
}

func (d *AllBridgeRestClient) ProtocolName() string {
	return d.name
}

func (d *AllBridgeRestClient) Get(ctx context.Context, from, to time.Time) (ProtocolActivity[Activity], error) {
	decoratedLogger := d.logger

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		decoratedLogger.Error("failed creating http request for retrieving protocol Activities",
			zap.Error(err),
		)
		return ProtocolActivity[Activity]{}, errors.WithStack(err)
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
		return ProtocolActivity[Activity]{}, errors.WithStack(err)
	}
	defer resp.Body.Close()

	decoratedLogger = decoratedLogger.
		With(zap.String("status_code", http.StatusText(resp.StatusCode))).
		With(zap.String("response_headers", commons.ToJson(resp.Header)))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		decoratedLogger.Error("error retrieving protocol Activities: got an invalid response status code",
			zap.String("response_body", string(body)),
		)
		return ProtocolActivity[Activity]{}, errors.Errorf("failed retrieving protocol Activities from url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err))
		return ProtocolActivity[Activity]{}, errors.Wrapf(errors.WithStack(err), "failed reading response body from protocol Activities. url:%s - status_code:%d", d.url, resp.StatusCode)
	}

	var temp allBridgeActivity
	err = json.Unmarshal(body, &temp)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return ProtocolActivity[Activity]{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol Activities. url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}

	return temp.toProtocolActivity()
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

func (m *allBridgeActivity) toProtocolActivity() (ProtocolActivity[Activity], error) {
	result := ProtocolActivity[Activity]{}

	totalValueSecured, err := strconv.ParseFloat(m.TotalValueSecured, 64)
	if err != nil {
		return result, errors.Wrap(err, "failed parsing string TotalValueSecure to float64")
	}
	result.TotalValueSecure = totalValueSecured

	totalValueTransferred, err := strconv.ParseFloat(m.TotalValueTransferred, 64)
	if err != nil {
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
