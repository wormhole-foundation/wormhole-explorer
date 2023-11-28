package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type cosmosRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Query string `json:"query"`
		Page  string `json:"page"`
	} `json:"params"`
}

type cosmosTxSearchParams struct {
	Sequence   string
	Timestamp  string
	SrcChannel string
	DstChannel string
}

type cosmosTxSearchResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Txs []struct {
			Hash     string `json:"hash"`
			Height   string `json:"height"`
			Index    int    `json:"index"`
			TxResult struct {
				Code      int    `json:"code"`
				Data      string `json:"data"`
				Log       string `json:"log"`
				Info      string `json:"info"`
				GasWanted string `json:"gas_wanted"`
				GasUsed   string `json:"gas_used"`
				Events    []struct {
					Type       string `json:"type"`
					Attributes []struct {
						Key   string `json:"key"`
						Value string `json:"value"`
						Index bool   `json:"index"`
					} `json:"attributes"`
				} `json:"events"`
				Codespace string `json:"codespace"`
			} `json:"tx_result"`
			Tx string `json:"tx"`
		} `json:"txs"`
		TotalCount string `json:"total_count"`
	} `json:"result"`
}

type cosmosEventResponse struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}

type cosmosLogWrapperResponse struct {
	Events []cosmosEventResponse `json:"events"`
}

type txSearchExtractor[T any] func(tx *cosmosTxSearchResponse, log []cosmosLogWrapperResponse) (T, error)

func fetchTxSearch[T any](ctx context.Context, baseUrl string, rl *time.Ticker, p *cosmosTxSearchParams, extractor txSearchExtractor[*T]) (*T, error) {
	queryTemplate := `send_packet.packet_sequence='%s' AND send_packet.packet_timeout_timestamp='%s' AND send_packet.packet_src_channel='%s' AND send_packet.packet_dst_channel='%s'`
	query := fmt.Sprintf(queryTemplate, p.Sequence, p.Timestamp, p.SrcChannel, p.DstChannel)
	q := cosmosRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "tx_search",
		Params: struct {
			Query string `json:"query"`
			Page  string `json:"page"`
		}{
			Query: query,
			Page:  "1",
		},
	}
	response, err := httpPost(ctx, rl, baseUrl, q)
	if err != nil {
		return nil, err
	}

	return parseTxSearchResponse[T](response, p, extractor)
}

func parseTxSearchResponse[T any](body []byte, p *cosmosTxSearchParams, extractor txSearchExtractor[*T]) (*T, error) {
	var txSearchReponse cosmosTxSearchResponse
	err := json.Unmarshal(body, &txSearchReponse)
	if err != nil {
		return nil, err
	}

	if len(txSearchReponse.Result.Txs) == 0 {
		return nil, fmt.Errorf("can not found hash for sequence %s, timestamp %s, srcChannel %s, dstChannel %s", p.Sequence, p.Timestamp, p.SrcChannel, p.DstChannel)
	}

	var log []cosmosLogWrapperResponse
	err = json.Unmarshal([]byte(txSearchReponse.Result.Txs[0].TxResult.Log), &log)
	if err != nil {
		return nil, err
	}

	return extractor(&txSearchReponse, log)
}
