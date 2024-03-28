package chains

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type apiWormchain struct {
	p2pNetwork    string
	evmosPool     *pool.Pool
	kujiraPool    *pool.Pool
	osmosisPool   *pool.Pool
	injectivePool *pool.Pool
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

type wormchainTx struct {
	srcChannel, dstChannel, sender, receiver, timestamp, sequence string
}

func fetchWormchainDetail(ctx context.Context, baseUrl string, txHash string) (*wormchainTx, error) {
	uri := fmt.Sprintf("%s/tx?hash=%s", baseUrl, txHash)
	body, err := httpGet(ctx, uri)
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
	return &wormchainTx{
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

func (a *apiWormchain) fetchOsmosisDetail(ctx context.Context, pool *pool.Pool, sequence, timestamp, srcChannel, dstChannel string, metrics metrics.Metrics) (*osmosisTx, error) {
	if pool == nil {
		return nil, fmt.Errorf("osmosis rpc pool not found")
	}

	osmosisRpcs := pool.GetItems()
	if len(osmosisRpcs) == 0 {
		return nil, fmt.Errorf("osmosis rpcs not found")
	}

	for _, rpc := range osmosisRpcs {
		rpc.Wait(ctx)
		osmosisTx, err := fetchOsmosisDetail(ctx, rpc.Id, sequence, timestamp, srcChannel, dstChannel)
		if osmosisTx != nil {
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDOsmosis), rpc.Description)
			return osmosisTx, nil
		}
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDOsmosis), rpc.Description)
		}
	}

	return nil, fmt.Errorf("osmosis tx not found")
}

func fetchOsmosisDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*osmosisTx, error) {
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

	response, err := httpPost(ctx, baseUrl, q)
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

func (a *apiWormchain) fetchEvmosDetail(ctx context.Context, pool *pool.Pool, sequence, timestamp, srcChannel, dstChannel string, metrics metrics.Metrics) (*evmosTx, error) {
	if pool == nil {
		return nil, fmt.Errorf("evmos rpc pool not found")
	}
	evmosRpcs := pool.GetItems()
	if len(evmosRpcs) == 0 {
		return nil, fmt.Errorf("evmos rpcs not found")
	}

	for _, rpc := range evmosRpcs {
		rpc.Wait(ctx)
		evmosTx, err := fetchEvmosDetail(ctx, rpc.Id, sequence, timestamp, srcChannel, dstChannel)
		if evmosTx != nil {
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDEvmos), rpc.Description)
			return evmosTx, nil
		}
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDEvmos), rpc.Description)
		}
	}
	return nil, fmt.Errorf("evmos tx not found")

}

func fetchEvmosDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*evmosTx, error) {
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

	//response, err := httpPost(ctx, rateLimiter, baseUrl, q)
	response, err := httpPost(ctx, baseUrl, q)
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

func (a *apiWormchain) fetchKujiraDetail(ctx context.Context, pool *pool.Pool, sequence, timestamp, srcChannel, dstChannel string, metrics metrics.Metrics) (*kujiraTx, error) {
	if pool == nil {
		return nil, fmt.Errorf("kujira rpc pool not found")
	}
	kujiraRpcs := pool.GetItems()
	if len(kujiraRpcs) == 0 {
		return nil, fmt.Errorf("kujira rpcs not found")
	}
	for _, rpc := range kujiraRpcs {
		rpc.Wait(ctx)
		kujiraTx, err := fetchKujiraDetail(ctx, rpc.Id, sequence, timestamp, srcChannel, dstChannel)
		if kujiraTx != nil {
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDKujira), rpc.Description)
			return kujiraTx, nil
		}
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDKujira), rpc.Description)
		}
	}
	return nil, fmt.Errorf("kujira tx not found")
}

func fetchKujiraDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*kujiraTx, error) {
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

	response, err := httpPost(ctx, baseUrl, q)
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

type injectiveRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Query string `json:"query"`
		Page  string `json:"page"`
	} `json:"params"`
}

type injectiveResponse struct {
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

type injectiveTx struct {
	txHash string
}

func (a *apiWormchain) fetchInjectiveDetail(ctx context.Context, pool *pool.Pool, sequence, timestamp, srcChannel, dstChannel string, metrics metrics.Metrics) (*injectiveTx, error) {
	if pool == nil {
		return nil, fmt.Errorf("injective rpc pool not found")
	}
	injectiveRpcs := pool.GetItems()
	if len(injectiveRpcs) == 0 {
		return nil, fmt.Errorf("injective rpcs not found")
	}
	for _, rpc := range injectiveRpcs {
		rpc.Wait(ctx)
		injectiveTx, err := fetchInjectiveDetail(ctx, rpc.Id, sequence, timestamp, srcChannel, dstChannel)
		if injectiveTx != nil {
			success := fmt.Sprintf("Successfully fetched transaction from injective: %s", rpc.Id)
			fmt.Sprintln(success)
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDInjective), rpc.Description)
			return injectiveTx, nil
		}
		error := fmt.Sprintf("Failed to fetch transaction from injective: %s", rpc.Id)
		fmt.Sprintln(error)
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDInjective), rpc.Description)
		}
	}
	return nil, fmt.Errorf("injective tx not found")
}

func fetchInjectiveDetail(ctx context.Context, baseUrl string, sequence, timestamp, srcChannel, dstChannel string) (*injectiveTx, error) {
	queryTemplate := `send_packet.packet_sequence='%s' AND send_packet.packet_timeout_timestamp='%s' AND send_packet.packet_src_channel='%s' AND send_packet.packet_dst_channel='%s'`
	query := fmt.Sprintf(queryTemplate, sequence, timestamp, srcChannel, dstChannel)
	q := injectiveRequest{
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

	response, err := httpPost(ctx, baseUrl, q)
	if err != nil {
		return nil, err
	}

	var iReponse injectiveResponse
	err = json.Unmarshal(response, &iReponse)
	if err != nil {
		return nil, err
	}

	if len(iReponse.Result.Txs) == 0 {
		return nil, fmt.Errorf("can not found hash for sequence %s, timestamp %s, srcChannel %s, dstChannel %s", sequence, timestamp, srcChannel, dstChannel)
	}
	return &injectiveTx{txHash: strings.ToLower(iReponse.Result.Txs[0].Hash)}, nil
}

type WorchainAttributeTxDetail struct {
	OriginChainID sdk.ChainID `bson:"originChainId"`
	OriginTxHash  string      `bson:"originTxHash"`
	OriginAddress string      `bson:"originAddress"`
}

func (a *apiWormchain) FetchWormchainTx(
	ctx context.Context,
	wormchainPool *pool.Pool,
	txHash string,
	metrics metrics.Metrics,
	logger *zap.Logger,
) (*TxDetail, error) {

	txHash = txHashLowerCaseWith0x(txHash)

	// Get the wormchain rpcs sorted by availability.
	wormchainRpcs := wormchainPool.GetItems()
	if len(wormchainRpcs) == 0 {
		return nil, errors.New("wormchain rpc pool is empty")
	}

	var wormchainTx *wormchainTx
	var err error
	for _, rpc := range wormchainRpcs {
		// wait for the rpc to be available
		rpc.Wait(ctx)
		wormchainTx, err = fetchWormchainDetail(ctx, rpc.Id, txHash)
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDWormchain), rpc.Description)
			logger.Debug("Failed to fetch transaction from wormchain", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		metrics.IncCallRpcSuccess(uint16(sdk.ChainIDWormchain), rpc.Description)
		break
	}

	if err != nil {
		return nil, err
	}
	if wormchainTx == nil {
		return nil, errors.New("failed to fetch wormchain transaction details")
	}

	// Verify if this transaction is from osmosis by wormchain
	if a.isOsmosisTx(wormchainTx) {
		osmosisTx, err := a.fetchOsmosisDetail(ctx, a.osmosisPool, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel, metrics)
		if err != nil {
			return nil, err
		}

		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: sdk.ChainIDOsmosis,
					OriginTxHash:  osmosisTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// Verify if this transaction is from kujira by wormchain
	if a.isKujiraTx(wormchainTx) {
		kujiraTx, err := a.fetchKujiraDetail(ctx, a.kujiraPool, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel, metrics)
		if err != nil {
			return nil, err
		}

		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: sdk.ChainIDKujira,
					OriginTxHash:  kujiraTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// Verify if this transaction is from evmos by wormchain
	if a.isEvmosTx(wormchainTx) {
		evmosTx, err := a.fetchEvmosDetail(ctx, a.evmosPool, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel, metrics)
		if err != nil {
			return nil, err
		}

		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: sdk.ChainIDEvmos,
					OriginTxHash:  evmosTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// Verify if this transaction is from injective by wormchain
	if a.isInjectiveTx(wormchainTx) {
		injectiveTx, err := a.fetchInjectiveDetail(ctx, a.injectivePool, wormchainTx.sequence, wormchainTx.timestamp, wormchainTx.srcChannel, wormchainTx.dstChannel, metrics)
		if err != nil {
			return nil, err
		}

		return &TxDetail{
			NativeTxHash: txHash,
			From:         wormchainTx.receiver,
			Attribute: &AttributeTxDetail{
				Type: "wormchain-gateway",
				Value: &WorchainAttributeTxDetail{
					OriginChainID: sdk.ChainIDInjective,
					OriginTxHash:  injectiveTx.txHash,
					OriginAddress: wormchainTx.sender,
				},
			},
		}, nil
	}

	// If the transaction is not from any known cosmos chain, increment the unknown wormchain transaction metric.
	metrics.IncWormchainUnknown(wormchainTx.srcChannel, wormchainTx.dstChannel)
	logger.Debug("Unknown wormchain transaction",
		zap.String("srcChannel", wormchainTx.srcChannel),
		zap.String("dstChannel", wormchainTx.dstChannel),
		zap.String("txHash", txHash))

	return &TxDetail{
		NativeTxHash: txHash,
		From:         wormchainTx.receiver,
	}, nil
}

func (a *apiWormchain) isOsmosisTx(tx *wormchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-2186" && tx.dstChannel == "channel-3"
	}
	if a.p2pNetwork == domain.P2pTestNet {
		return tx.srcChannel == "channel-3086" && tx.dstChannel == "channel-5"
	}
	return false
}

func (a *apiWormchain) isKujiraTx(tx *wormchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-113" && tx.dstChannel == "channel-9"
	}
	// Pending get channels for testnet
	// if a.p2pNetwork == domain.P2pTestNet {
	// 	return tx.srcChannel == "" && tx.dstChannel == ""
	// }
	return false
}

func (a *apiWormchain) isEvmosTx(tx *wormchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-94" && tx.dstChannel == "channel-5"
	}
	// Pending get channels for testnet
	// if a.p2pNetwork == domain.P2pTestNet {
	// 	return tx.srcChannel == "" && tx.dstChannel == ""
	// }
	return false
}

func (a *apiWormchain) isInjectiveTx(tx *wormchainTx) bool {
	if a.p2pNetwork == domain.P2pMainNet {
		return tx.srcChannel == "channel-183" && tx.dstChannel == "channel-13"
	}
	// Pending get channels for testnet
	// if a.p2pNetwork == domain.P2pTestNet {
	// 	return tx.srcChannel == "" && tx.dstChannel == ""
	// }
	return false
}
