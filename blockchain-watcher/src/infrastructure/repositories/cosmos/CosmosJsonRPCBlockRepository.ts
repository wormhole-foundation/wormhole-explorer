import { JsonRPCBlockRepositoryCfg, ProviderPoolMap } from "../RepositoriesBuilder";
import { CosmosTransaction } from "../../../domain/entities/cosmos";
import { CosmosRepository } from "../../../domain/repositories";
import { getChainProvider } from "../common/utils";
import { Filter } from "../../../domain/actions/cosmos/types";
import winston from "winston";
import { ProviderHealthCheck } from "../../../domain/actions/poolRpcs/PoolRpcs";

const TRANSACTION_SEARCH_ENDPOINT = "/tx_search";
const BLOCK_ENDPOINT = "/block";

export class CosmosJsonRPCBlockRepository implements CosmosRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;
  protected cfg: JsonRPCBlockRepositoryCfg;

  constructor(cfg: JsonRPCBlockRepositoryCfg, pool: ProviderPoolMap) {
    this.logger = winston.child({ module: "CosmosJsonRPCBlockRepository" });
    this.pool = pool;
    this.cfg = cfg;
  }

  async healthCheck(chain: string, finality: string, cursor: bigint): Promise<void> {
    const resultHeight: ProviderHealthCheck[] = [];
    const pool = this.pool[chain];
    const providers = pool.getProviders();
    const blockEndpoint = `${BLOCK_ENDPOINT}?height=${cursor}`;
    let resultsBlock: ResultBlock;

    for (const provider of providers) {
      try {
        const requestStartTime = performance.now();
        resultsBlock = await provider.get<typeof resultsBlock>(blockEndpoint);
        const requestEndTime = performance.now();

        const result = (
          "result" in resultsBlock ? resultsBlock.result : resultsBlock
        ) as ResultBlock;

        resultHeight.push({
          url: provider.getUrl(),
          height: BigInt(result.block.header.height),
          isLive: true,
          latency: Number(((requestEndTime - requestStartTime) / 1000).toFixed(2)),
        });
      } catch (e) {
        resultHeight.push({ url: provider.getUrl(), height: undefined, isLive: false });
      }
    }
    pool.setProviders(providers, resultHeight, cursor);
  }

  async getTransactions(
    filter: Filter,
    blockBatchSize: number,
    chain: string
  ): Promise<CosmosTransaction[]> {
    try {
      const cosmosTransaction = [];
      const query = `"wasm._contract_address='${filter.addresses[0]}'"`;
      let resultTransactionSearch: ResultTransactionSearch;
      let continuesFetching = true;
      let page = 1;

      while (continuesFetching) {
        try {
          resultTransactionSearch = await getChainProvider(chain, this.pool).get<
            typeof resultTransactionSearch
          >(
            `${TRANSACTION_SEARCH_ENDPOINT}?query=${query}&page=${page}&per_page=${blockBatchSize}`
          );

          // Dependes the chain, the result can be different. Sei dose not have a result key, Terra, Terra2 and Xpla containers a result key
          const result = (
            "result" in resultTransactionSearch
              ? resultTransactionSearch.result
              : resultTransactionSearch
          ) as ResultTransactionSearch;

          if (result && result.txs) {
            cosmosTransaction.push(...result.txs);

            if (result.txs.length < blockBatchSize || Number(result.total_count) < blockBatchSize) {
              continuesFetching = false;
            }
            page++;
          }

          if (result?.message === "Invalid request") {
            continuesFetching = false;
          }
        } catch (e) {
          this.handleError(`Error: ${e}`, "getTransactions", chain);
          continuesFetching = false;
        }
      }

      if (!cosmosTransaction) {
        this.logger.warn(
          `[getTransactions][${chain}] Do not find any transaction with query \n${query}\n`
        );
        return [];
      }

      const sortedCosmosTransaction = cosmosTransaction.sort(
        (a, b) => Number(a.height) - Number(b.height)
      );
      return sortedCosmosTransaction.map((tx) => {
        return {
          events: tx.tx_result.events,
          height: BigInt(tx.height),
          data: tx.tx_result.data,
          hash: tx.hash,
          tx: Buffer.from(tx.tx, "base64"),
          chain,
        };
      });
    } catch (e) {
      this.handleError(`Error: ${e}`, "getTransactions", chain);
      throw e;
    }
  }

  async getBlockTimestamp(blockNumber: bigint, chain: string): Promise<number | undefined> {
    try {
      const blockEndpoint = `${BLOCK_ENDPOINT}?height=${blockNumber}`;
      let resultsBlock: ResultBlock;

      resultsBlock = await getChainProvider(chain, this.pool).get<typeof resultsBlock>(
        blockEndpoint
      );

      const result = ("result" in resultsBlock ? resultsBlock.result : resultsBlock) as ResultBlock;

      if (!result || !result.block || !result.block.header || !result.block.header.time) {
        return undefined;
      }

      const dateTime: Date = new Date(result.block.header.time);
      const timestamp: number = Math.floor(dateTime.getTime() / 1000);

      return timestamp;
    } catch (e: Error | any) {
      if (e.toString().includes("undefined")) {
        return undefined;
      }
      this.handleError(`Error: ${e}`, "getBlockTimestamp", chain);
      throw e;
    }
  }

  private handleError(e: any, method: string, chain: string) {
    this.logger.error(`[${chain}] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultTransactionSearch = {
  total_count: string;
  message: string;
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
      tx: string;
    }
  ];
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

type ResultBlock = {
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
