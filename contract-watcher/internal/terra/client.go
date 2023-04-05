package terra

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/ratelimit"
)

// TerraSDK is a client for the Terra blockchain.
type TerraSDK struct {
	url    string
	client *http.Client
	rl     ratelimit.Limiter
}

// TerraTrx is a transaction on the Terra blockchain.
type TerraTrx struct {
}

// NewTerraSDK creates a new TerraSDK.
func NewTerraSDK(url string, rl ratelimit.Limiter) *TerraSDK {
	return &TerraSDK{
		url:    url,
		rl:     rl,
		client: &http.Client{},
	}
}

type LastBlockResponse struct {
	Block struct {
		Header struct {
			Height string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

// GetLastBlock returns the last block height.
func (t *TerraSDK) GetLastBlock(ctx context.Context) (int64, error) {
	lastBlockURL := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", t.url)
	req, err := http.NewRequest(http.MethodGet, lastBlockURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")

	t.rl.Take()

	res, err := t.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var response LastBlockResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	lastBlockHeight, err := strconv.ParseInt(response.Block.Header.Height, 10, 64)
	if err != nil {
		return 0, err
	}

	return lastBlockHeight, nil
}

type TxByBlockResponse struct {
	Limit      int  `json:"limit"`
	NextOffset *int `json:"next"`
	Txs        []Tx `json:"txs"`
}

type Tx struct {
	ID        int        `json:"id"`
	Tx        any        `json:"tx"`
	Logs      any        `json:"logs"`
	Code      int        `json:"code"`
	Height    string     `json:"height"`
	Txhash    string     `json:"txhash"`
	RawLog    string     `json:"raw_log"`
	Timestamp *time.Time `json:"timestamp"`
}

type WormholeTerraTx struct {
	Type  string `json:"type"`
	Value struct {
		Fee struct {
			Gas    string `json:"gas"`
			Amount []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"amount"`
		} `json:"fee"`
		Msg []struct {
			Type  string `json:"type"`
			Value struct {
				Coins      []any  `json:"coins"`
				Sender     string `json:"sender"`
				Contract   string `json:"contract"`
				ExecuteMsg struct {
					SubmitVaa struct {
						Data []byte `json:"data"`
					} `json:"submit_vaa"`
				} `json:"execute_msg"`
			} `json:"value"`
		} `json:"msg"`
		Memo       string `json:"memo"`
		Signatures []struct {
			PubKey struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			Signature string `json:"signature"`
		} `json:"signatures"`
		TimeoutHeight string `json:"timeout_height"`
	} `json:"value"`
}

type WormholeTerraTxLog struct {
	Log struct {
		Tax string `json:"tax"`
	} `json:"log"`
	Events []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"events"`
	MsgIndex int `json:"msg_index"`
}

// GetTransactionsByBlockHeight returns the transactions for a given block height.
func (t *TerraSDK) GetTransactionsByBlockHeight(ctx context.Context, height int64, offset *int) (*TxByBlockResponse, error) {
	transactionsByBlockURL := fmt.Sprintf("%s/v1/txs", t.url)
	req, err := http.NewRequest(http.MethodGet, transactionsByBlockURL, nil)
	if err != nil {
		return nil, err
	}
	values := req.URL.Query()
	values.Add("block", strconv.FormatInt(height, 10))
	values.Add("limit", "100")
	if offset != nil {
		values.Add("offset", strconv.Itoa(*offset))
	}
	req.URL.RawQuery = values.Encode()

	req.Header.Add("Content-Type", "application/json")

	t.rl.Take()

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response TxByBlockResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
