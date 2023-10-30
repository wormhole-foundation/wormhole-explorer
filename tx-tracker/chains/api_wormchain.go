package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type apiWormchain struct {
	osmosisUrl         string
	osmosisRateLimiter *time.Ticker
	kujiraUrl          string
	kujiraRateLimiter  *time.Ticker
	evmosUrl           string
	evmosRateLimiter   *time.Ticker
	p2pNetwork         string
}

type wormchainTxDetail struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
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
	} `json:"result"`
}

type event struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}

type packetData struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

type logWrapper struct {
	Events []event `json:"events"`
}

type worchainTx struct {
	srcChannel, dstChannel, sender, receiver, timestamp, sequence string
}

func fetchWormchainDetail(ctx context.Context, baseUrl string, rateLimiter *time.Ticker, txHash string) (*worchainTx, error) {
	uri := fmt.Sprintf("%s/tx?hash=%s", baseUrl, txHash)
	body, err := httpGet(ctx, rateLimiter, uri)
	if err != nil {
		return nil, err
	}

	var tx wormchainTxDetail
	err = json.Unmarshal(body, &tx)
	if err != nil {
		return nil, err
	}

	var log []logWrapper
	err = json.Unmarshal([]byte(tx.Result.TxResult.Log), &log)
	if err != nil {
		return nil, err
	}

	var srcChannel, dstChannel, sender, receiver, timestamp, sequence string
	for _, l := range log {
		for _, e := range l.Events {
			if e.Type == "recv_packet" {
				for _, attr := range e.Attributes {
					if attr.Key == "packet_src_channel" {
						srcChannel = attr.Value
					}
					if attr.Key == "packet_dst_channel" {
						dstChannel = attr.Value
					}
					if attr.Key == "packet_timeout_timestamp" {
						timestamp = attr.Value
					}

					if attr.Key == "packet_sequence" {
						sequence = attr.Value
					}

					if attr.Key == "packet_data" {
						var pd packetData
						err = json.Unmarshal([]byte(attr.Value), &pd)
						if err != nil {
							return nil, err
						}
						sender = pd.Sender
						receiver = pd.Receiver
					}
				}
			}
		}
	}
	return &worchainTx{
		srcChannel: srcChannel,
		dstChannel: dstChannel,
		sender:     sender,
		receiver:   receiver,
		timestamp:  timestamp,
		sequence:   sequence,
	}, nil

}

type osmosisRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Query string `json:"query"`
		Page  string `json:"page"`
	} `json:"params"`
}

type osmosisResponse struct {
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

type osmosisTx struct {
	txHash string
}

func fetchOsmosisDetail(ctx context.Context, baseUrl string, rateLimiter *time.Ticker, sequence, timestamp, srcChannel, dstChannel string) (*osmosisTx, error) {
	queryTemplate := `send_packet.packet_sequence='%s' AND send_packet.packet_timeout_timestamp='%s' AND send_packet.packet_src_channel='%s' AND send_packet.packet_dst_channel='%s'`
	query := fmt.Sprintf(queryTemplate, sequence, timestamp, srcChannel, dstChannel)
	q := osmosisRequest{
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

	response, err := httpPost(ctx, rateLimiter, baseUrl, q)
	if err != nil {
		return nil, err
	}

	var oReponse osmosisResponse
	err = json.Unmarshal(response, &oReponse)
	if err != nil {
		return nil, err
	}

	if len(oReponse.Result.Txs) == 0 {
		return nil, fmt.Errorf("can not found hash for sequence %s, timestamp %s, srcChannel %s, dstChannel %s", sequence, timestamp, srcChannel, dstChannel)
	}
	return &osmosisTx{txHash: strings.ToLower(oReponse.Result.Txs[0].Hash)}, nil
}

type evmosRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Query string `json:"query"`
		Page  string `json:"page"`
	} `json:"params"`
}

type evmosResponse struct {
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

type evmosTx struct {
	txHash string
}

func fetchEvmosDetail(ctx context.Context, baseUrl string, rateLimiter *time.Ticker, sequence, timestamp, srcChannel, dstChannel string) (*evmosTx, error) {
	queryTemplate := `send_packet.packet_sequence='%s' AND send_packet.packet_timeout_timestamp='%s' AND send_packet.packet_src_channel='%s' AND send_packet.packet_dst_channel='%s'`
	query := fmt.Sprintf(queryTemplate, sequence, timestamp, srcChannel, dstChannel)
	q := evmosRequest{
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

	response, err := httpPost(ctx, rateLimiter, baseUrl, q)
	if err != nil {
		return nil, err
	}

	var eReponse evmosResponse
	err = json.Unmarshal(response, &eReponse)
	if err != nil {
		return nil, err
	}

	if len(eReponse.Result.Txs) == 0 {
		return nil, fmt.Errorf("can not found hash for sequence %s, timestamp %s, srcChannel %s, dstChannel %s", sequence, timestamp, srcChannel, dstChannel)
	}
	return &evmosTx{txHash: strings.ToLower(eReponse.Result.Txs[0].Hash)}, nil
}

type kujiraRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Query string `json:"query"`
		Page  string `json:"page"`
	} `json:"params"`
}

type kujiraResponse struct {
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

type kujiraTx struct {
	txHash string
}

func fetchKujiraDetail(ctx context.Context, baseUrl string, rateLimiter *time.Ticker, sequence, timestamp, srcChannel, dstChannel string) (*kujiraTx, error) {
	queryTemplate := `send_packet.packet_sequence='%s' AND send_packet.packet_timeout_timestamp='%s' AND send_packet.packet_src_channel='%s' AND send_packet.packet_dst_channel='%s'`
	query := fmt.Sprintf(queryTemplate, sequence, timestamp, srcChannel, dstChannel)
	q := kujiraRequest{
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

	response, err := httpPost(ctx, rateLimiter, baseUrl, q)
	if err != nil {
		return nil, err
	}

	var kReponse kujiraResponse
	err = json.Unmarshal(response, &kReponse)
	if err != nil {
		return nil, err
	}

	if len(kReponse.Result.Txs) == 0 {
		return nil, fmt.Errorf("can not found hash for sequence %s, timestamp %s, srcChannel %s, dstChannel %s", sequence, timestamp, srcChannel, dstChannel)
	}
	return &kujiraTx{txHash: strings.ToLower(kReponse.Result.Txs[0].Hash)}, nil
}

type WorchainAttributeTxDetail struct {
	OriginChainID sdk.ChainID `bson:"originChainId"`
	OriginTxHash  string      `bson:"originTxHash"`
	OriginAddress string      `bson:"originAddress"`
}

func (a *apiWormchain) fetchWormchainTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	txHash = txHashLowerCaseWith0x(txHash)

	wormchainTx, err := fetchWormchainDetail(ctx, baseUrl, rateLimiter, txHash)
	if err != nil {
		return nil, err
	}

	// Verify if this transaction is from osmosis by wormchain
	if a.isOsmosisTx(wormchainTx) {
		osmosisTx, err := fetchOsmosisDetail(ctx, a.osmosisUrl, a.osmosisRateLimiter, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
		if err != nil {
			return nil, err
		}
		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: ChainIDOsmosis,
					OriginTxHash:  osmosisTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// Verify if this transaction is from kujira by wormchain
	if a.isKujiraTx(wormchainTx) {
		kujiraTx, err := fetchKujiraDetail(ctx, a.kujiraUrl, a.kujiraRateLimiter, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
		if err != nil {
			return nil, err
		}
		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: ChainIDKujira,
					OriginTxHash:  kujiraTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// Verify if this transaction is from evmos by wormchain
	if a.isEvmosTx(wormchainTx) {
		evmosTx, err := fetchEvmosDetail(ctx, a.evmosUrl, a.evmosRateLimiter, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel)
		if err != nil {
			return nil, err
		}
		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: ChainIDEvmos,
					OriginTxHash:  evmosTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	return &TxDetail{
		NativeTxHash: txHash,
		From:         wormchainTx.receiver,
	}, nil
}

func (a *apiWormchain) isOsmosisTx(tx *worchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-2186" && tx.dstChannel == "channel-3"
	}
	if a.p2pNetwork == domain.P2pTestNet {
		return tx.srcChannel == "channel-3086" && tx.dstChannel == "channel-5"
	}
	return false
}

func (a *apiWormchain) isKujiraTx(tx *worchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-113" && tx.dstChannel == "channel-9"
	}
	// Pending get channels for testnet
	// if a.p2pNetwork == domain.P2pTestNet {
	// 	return tx.srcChannel == "" && tx.dstChannel == ""
	// }
	return false
}

func (a *apiWormchain) isEvmosTx(tx *worchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-94" && tx.dstChannel == "channel-5"
	}
	// Pending get channels for testnet
	// if a.p2pNetwork == domain.P2pTestNet {
	// 	return tx.srcChannel == "" && tx.dstChannel == ""
	// }
	return false
}
