import { divideIntoBatches, getChainProvider, hexToHash } from "../common/utils";
import { ProviderPoolMap, JsonRPCBlockRepositoryCfg } from "../RepositoriesBuilder";
import { WormchainRepository, ProviderHealthCheck } from "../../../domain/repositories";
import { setTimeout } from "timers/promises";
import { mapChain } from "../../../common/wormchain";
import winston from "winston";
import {
  WormchainTransaction,
  WormchainBlockLogs,
  IbcTransaction,
  CosmosRedeem,
} from "../../../domain/entities/wormchain";

let TRANSACTION_SEARCH_ENDPOINT = "/tx_search";
let BLOCK_HEIGHT_ENDPOINT = "/abci_info";
let TRANSACTION_ENDPOINT = "/tx";
let BLOCK_ENDPOINT = "/block";

const GROW_SLEEP_TIME = 350;
const MAX_ATTEMPTS = 20;

export class WormchainJsonRPCBlockRepository implements WormchainRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;
  protected cfg: JsonRPCBlockRepositoryCfg;

  constructor(cfg: JsonRPCBlockRepositoryCfg, pool: ProviderPoolMap) {
    this.logger = winston.child({ module: "WormchainJsonRPCBlockRepository" });
    this.pool = pool;
    this.cfg = cfg;
  }

  async healthCheck(
    chain: string,
    finality: string,
    cursor: bigint
  ): Promise<ProviderHealthCheck[]> {
    const providersHealthCheck: ProviderHealthCheck[] = [];
    const pool = this.pool[chain];
    const providers = pool.getProviders();

    for (const provider of providers) {
      try {
        const results = await provider.get<ResultBlockHeight>(BLOCK_HEIGHT_ENDPOINT);

        if (results?.result?.response?.last_block_height) {
          const height = results.result.response
            ? BigInt(results.result.response.last_block_height)
            : undefined;
          providersHealthCheck.push({
            isHealthy: height !== undefined,
            latency: provider.getLatency(),
            height: height,
            url: provider.getUrl(),
          });
        }
      } catch (e) {
        this.logger.error(
          `[${chain}][healthCheck] Error getting result on ${provider.getUrl()}: ${JSON.stringify(
            e
          )}`
        );
        providersHealthCheck.push({ url: provider.getUrl(), height: undefined, isHealthy: false });
      }
    }
    pool.setProviders(chain, providers, providersHealthCheck, cursor);
    return providersHealthCheck;
  }

  async getBlockHeight(chain: string): Promise<bigint | undefined> {
    try {
      const results = await getChainProvider(chain, this.pool).get<ResultBlockHeight>(
        BLOCK_HEIGHT_ENDPOINT
      );

      if (results?.result?.response?.last_block_height) {
        const blockHeight = results.result.response.last_block_height;
        return BigInt(blockHeight);
      }
      return undefined;
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  async getBlockLogs(
    chain: string,
    blockNumber: bigint,
    attributesTypes: string[]
  ): Promise<WormchainBlockLogs> {
    try {
      const blockEndpoint = `${BLOCK_ENDPOINT}?height=${blockNumber}`;
      // Set up cosmos client
      const cosmosClient = getChainProvider(chain, this.pool);

      // Get wormchain block data
      const resultsBlock = await cosmosClient.get<ResultBlock>(blockEndpoint);
      const txs = resultsBlock.result.block.data.txs;

      if (!txs || txs.length === 0) {
        return {
          transactions: [],
          blockHeight: BigInt(resultsBlock.result.block.header.height),
          timestamp: Number(resultsBlock.result.block.header.time),
        };
      }

      const wormchainTransactions: WormchainTransaction[] = [];

      const hashNumbers = new Set(txs.map((tx) => tx));
      const batches = divideIntoBatches(hashNumbers, 10);

      for (const batch of batches) {
        for (const hashBatch of batch) {
          const hash: string = hexToHash(hashBatch);
          const txEndpoint = `${TRANSACTION_ENDPOINT}?hash=0x${hash}`;

          // Get wormchain transactions data
          const resultTransaction: ResultTransaction = await cosmosClient.get<
            typeof resultTransaction
          >(txEndpoint);

          if (
            resultTransaction &&
            resultTransaction.result.tx_result &&
            resultTransaction.result.tx_result.events
          ) {
            const groupedAttributes: {
              key: string;
              value: string;
              index: boolean;
            }[] = [];

            // Group all attributes by tx hash
            resultTransaction.result.tx_result.events
              .filter((event) => attributesTypes.includes(event.type))
              .forEach((event) => {
                event.attributes.forEach((attr) => {
                  groupedAttributes.push(attr);
                });
              });

            if (groupedAttributes && groupedAttributes.length > 0) {
              const txToBase64 = Buffer.from(resultTransaction.result.tx, "base64");

              wormchainTransactions.push({
                attributes: groupedAttributes,
                height: BigInt(resultTransaction.result.height),
                hash: `0x${resultTransaction.result.hash}`.toLocaleLowerCase(),
                tx: txToBase64,
              });
            }
          }
        }
      }
      const dateTime: Date = new Date(resultsBlock.result.block.header.time);
      const timestamp: number = Math.floor(dateTime.getTime() / 1000);

      return {
        transactions: wormchainTransactions || [],
        blockHeight: BigInt(resultsBlock.result.block.header.height),
        timestamp,
      };
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  async getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]> {
    try {
      // Set up cosmos client
      const chain = mapChain(ibcTransaction.targetChain);
      const cosmosClient = getChainProvider(chain, this.pool);

      let resultTransactionSearch: ResultTransactionSearch | undefined;
      let isIBCTransferFinalized = false;
      let sleepTime = 300;
      let attempts = 1;

      const query = `"recv_packet.packet_sequence=${ibcTransaction.sequence} AND 
          recv_packet.packet_timeout_timestamp='${ibcTransaction.timestamp}' AND 
          recv_packet.packet_src_channel='${ibcTransaction.srcChannel}' AND 
          recv_packet.packet_dst_channel='${ibcTransaction.dstChannel}'"`;

      // The process to find the reedeem on target chain sometimes takes a while so we need to wait for it to be finalized before returning the data
      // we will try to get the transaction data every 300ms (and increasing) until it is finalized if it takes more than 10 attempts, we will throw an error
      while (!isIBCTransferFinalized && attempts <= MAX_ATTEMPTS) {
        try {
          await this.sleep(sleepTime);

          // Get cosmos transactions data
          resultTransactionSearch = await cosmosClient.get<typeof resultTransactionSearch>(
            `${TRANSACTION_SEARCH_ENDPOINT}?query=${query}&prove=false&page=1&per_page=1`
          );

          if (
            resultTransactionSearch &&
            resultTransactionSearch.result &&
            resultTransactionSearch.result.txs &&
            resultTransactionSearch.result.txs.length > 0
          ) {
            isIBCTransferFinalized = true;
            break;
          }
        } catch (e) {
          this.handleError(
            `[${ibcTransaction.targetChain}] Get transaction error: ${e} with query \n${query}\n`,
            "getRedeems"
          );
        }

        this.logger.warn(
          `[getRedeems] Attempt ${attempts} to get transaction with chainId: ${ibcTransaction.targetChain}. Retrying in ${sleepTime}ms`
        );
        if (!isIBCTransferFinalized) {
          sleepTime += GROW_SLEEP_TIME;
          attempts++;
        }
      }

      if (
        !resultTransactionSearch ||
        !resultTransactionSearch.result ||
        !resultTransactionSearch.result.txs
      ) {
        return [];
      }

      return resultTransactionSearch.result.txs.map((tx) => {
        return {
          blockTimestamp: ibcTransaction.blockTimestamp,
          timestamp: ibcTransaction.timestamp,
          chainId: ibcTransaction.targetChain,
          events: tx.tx_result.events,
          height: BigInt(tx.height),
          data: tx.tx_result.data,
          hash: tx.hash,
          tx: ibcTransaction.tx,
        };
      });
    } catch (e) {
      this.handleError(`Error: ${e}`, "getRedeems");
      throw e;
    }
  }

  private async sleep(sleepTime: number) {
    await setTimeout(sleepTime, null, { ref: false });
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[wormchain] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultBlockHeight = {
  result: {
    response: {
      last_block_height: string;
    };
  };
};

type ResultBlock = {
  result: {
    block_id: {
      hash: string;
      parts: {
        total: number;
        hash: string;
      };
    };
    block: {
      header: {
        version: { block: string };
        chain_id: string;
        height: string;
        time: string; // eg. '2023-01-03T12:13:00.849094631Z'
        last_block_id: { hash: string; parts: { total: number; hash: string } };
        last_commit_hash: string;
        data_hash: string;
        validators_hash: string;
        next_validators_hash: string;
        consensus_hash: string;
        app_hash: string;
        last_results_hash: string;
        evidence_hash: string;
        proposer_address: string;
      };
      data: { txs: string[] | null };
      evidence: { evidence: null };
      last_commit: {
        height: string;
        round: number;
        block_id: { hash: string; parts: { total: number; hash: string } };
        signatures: string[];
      };
    };
  };
};

type ResultTransaction = {
  result: {
    height: string;
    hash: string;
    tx: string;
    tx_result: {
      height: string;
      txhash: string;
      codespace: string;
      code: 0;
      data: string;
      raw_log: string;
      logs: [{ msg_index: number; log: string; events: EventsType }];
      info: string;
      gas_wanted: string;
      gas_used: string;
      tx: {
        "@type": "/cosmos.tx.v1beta1.Tx";
        body: {
          messages: [
            {
              "@type": "/cosmos.staking.v1beta1.MsgBeginRedelegate";
              delegator_address: string;
              validator_src_address: string;
              validator_dst_address: string;
              amount: { denom: string; amount: string };
            }
          ];
          memo: "";
          timeout_height: "0";
          extension_options: [];
          non_critical_extension_options: [];
        };
        auth_info: {
          signer_infos: [
            {
              public_key: {
                "@type": "/cosmos.crypto.secp256k1.PubKey";
                key: string;
              };
              mode_info: { single: { mode: string } };
              sequence: string;
            }
          ];
          fee: {
            amount: [{ denom: string; amount: string }];
            gas_limit: string;
            payer: string;
            granter: string;
          };
        };
        signatures: string[];
      };
      timestamp: string; // eg. '2023-01-03T12:12:54Z'
      events: EventsType[];
    };
  };
};

type ResultTransactionSearch = {
  result: {
    txs: [
      {
        height: string;
        hash: string;
        tx_result: {
          height: string;
          txhash: string;
          codespace: string;
          code: 0;
          data: string;
          raw_log: string;
          logs: [{ msg_index: number; log: string; events: EventsType }];
          info: string;
          gas_wanted: string;
          gas_used: string;
          tx: {
            "@type": "/cosmos.tx.v1beta1.Tx";
            body: {
              messages: [
                {
                  "@type": "/cosmos.staking.v1beta1.MsgBeginRedelegate";
                  delegator_address: string;
                  validator_src_address: string;
                  validator_dst_address: string;
                  amount: { denom: string; amount: string };
                }
              ];
              memo: "";
              timeout_height: "0";
              extension_options: [];
              non_critical_extension_options: [];
            };
            auth_info: {
              signer_infos: [
                {
                  public_key: {
                    "@type": "/cosmos.crypto.secp256k1.PubKey";
                    key: string;
                  };
                  mode_info: { single: { mode: string } };
                  sequence: string;
                }
              ];
              fee: {
                amount: [{ denom: string; amount: string }];
                gas_limit: string;
                payer: string;
                granter: string;
              };
            };
            signatures: string[];
          };
          timestamp: string; // eg. '2023-01-03T12:12:54Z'
          events: EventsType[];
        };
      }
    ];
  };
};

type EventsType = {
  type: string;
  attributes: [
    {
      key: string;
      value: string;
      index: boolean;
    }
  ];
};
