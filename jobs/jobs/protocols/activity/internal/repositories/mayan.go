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

func NewMayanRestClient(name, url string, logger *zap.Logger, httpClient commons.HttpDo) *MayanRestClient {
	return &MayanRestClient{
		name:   name,
		url:    url,
		logger: logger,
		client: httpClient,
	}
}

type MayanRestClient struct {
	name   string
	url    string
	client commons.HttpDo
	logger *zap.Logger
}

func (d *MayanRestClient) ProtocolName() string {
	return d.name
}

func (d *MayanRestClient) Get(ctx context.Context, from, to time.Time) (ProtocolActivity[Activity], error) {
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

	type mayanActivity struct {
		AlternativeEmitterChainID string  `json:"emmiter_chain_id"` // typo is on purpose due to mayan-api returning in that format
		DestinationChainID        string  `json:"destination_chain_id"`
		Txs                       uint64  `json:"txs"`
		TotalUSD                  float64 `json:"total_usd"`
	}

	var mayanResp ProtocolActivity[mayanActivity]
	err = json.Unmarshal(body, &mayanResp)
	if err != nil {
		decoratedLogger.Error("failed reading response body", zap.Error(err), zap.String("response_body", string(body)))
		return ProtocolActivity[Activity]{}, errors.Wrapf(errors.WithStack(err), "failed unmarshalling response body from protocol Activities. url:%s - status_code:%d - response_body:%s", d.url, resp.StatusCode, string(body))
	}

	result := ProtocolActivity[Activity]{
		TotalValueTransferred: mayanResp.TotalValueTransferred,
		TotalValueSecure:      mayanResp.TotalValueSecure,
		TotalMessages:         mayanResp.TotalMessages,
		Volume:                mayanResp.Volume,
	}

	for _, act := range mayanResp.Activities {

		emitterChainId, errEmitter := strconv.ParseUint(act.AlternativeEmitterChainID, 10, 64)
		if errEmitter != nil {
			return ProtocolActivity[Activity]{}, errors.Wrap(errEmitter, "failed parsing protocol activity emitter chain id from string to uint64")
		}

		destChainId, errDest := strconv.ParseUint(act.DestinationChainID, 10, 64)
		if errDest != nil {
			return ProtocolActivity[Activity]{}, errors.Wrap(errDest, "failed parsing protocol activity destination chain id from string to uint64")
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
