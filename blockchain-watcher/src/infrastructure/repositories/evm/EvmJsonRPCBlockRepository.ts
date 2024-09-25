import { JsonRPCBlockRepositoryCfg, ProviderPoolMap } from "../RepositoriesBuilder";
import { divideIntoBatches, getChainProvider } from "../common/utils";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { ProviderPoolDecorator } from "../../rpc/http/ProviderPoolDecorator";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import { EvmBlockRepository } from "../../../domain/repositories";
import { HttpClientError } from "../../errors/HttpClientError";
import winston from "../../log";
import {
  ReceiptTransaction,
  EvmLogFilter,
  EvmBlock,
  EvmLog,
  EvmTag,
} from "../../../domain/entities";

/**
 * EvmJsonRPCBlockRepository is a repository that uses a JSON RPC endpoint to fetch blocks.
 * On the reliability side, only knows how to timeout.
 */

const HEXADECIMAL_PREFIX = "0x";
const TX_BATCH_SIZE = 10;

export class EvmJsonRPCBlockRepository implements EvmBlockRepository {
  protected pool: ProviderPoolMap;
  protected cfg: JsonRPCBlockRepositoryCfg;
  protected readonly logger;

  constructor(
    cfg: JsonRPCBlockRepositoryCfg,
    pool: Record<string, ProviderPoolDecorator<InstrumentedHttpProvider>>
  ) {
    this.cfg = cfg;
    this.pool = pool;
    this.logger = winston.child({ module: "EvmJsonRPCBlockRepository" });
    this.logger.info(`Created for ${Object.keys(this.cfg.chains)}`);
  }

  async healthCheck(
    chain: string,
    finality: EvmTag,
    cursor: bigint
  ): Promise<ProviderHealthCheck[]> {
    const pool = this.pool[chain];
    const providers = pool.getProviders();
    const providersHealthCheck: ProviderHealthCheck[] = [];

    for (const provider of providers) {
      try {
        const requestStartTime = performance.now();
        const response = await this.getBlockByNumber(
          provider,
          finality,
          false // isTransactionsPresent
        );
        const requestEndTime = performance.now();

        const height = response.result?.number ? BigInt(response.result.number) : undefined;

        providersHealthCheck.push({
          isHealthy: height !== undefined,
          latency: Number(((requestEndTime - requestStartTime) / 1000).toFixed(2)),
          height: height,
          url: provider.getUrl(),
        });
      } catch (e) {
        providersHealthCheck.push({ url: provider.getUrl(), height: undefined, isHealthy: false });
      }
    }
    pool.setProviders(chain, providers, providersHealthCheck, cursor);
    return providersHealthCheck;
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const block: EvmBlock = await this.getBlock(chain, finality);
    return block.number;
  }

  /**
   * Get blocks by block number.
   * @param blockNumbers
   * @returns a record of block hash -> EvmBlock
   */
  async getBlocks(
    chain: string,
    blockNumbers: Set<bigint>,
    isTransactionsPresent: boolean = false
  ): Promise<Record<string, EvmBlock>> {
    if (!blockNumbers.size) return {};

    let combinedResults: ResultBlocks[] = [];
    const provider = getChainProvider(chain, this.pool);
    const chainCfg = this.getCurrentChain(chain);
    const batches = divideIntoBatches(blockNumbers, 9);

    for (const batch of batches) {
      const reqs: any[] = [];
      for (let blockNumber of batch) {
        const blockNumberStrParam = `${HEXADECIMAL_PREFIX}${blockNumber.toString(16)}`;
        const blockNumberStrId = blockNumber.toString();

        reqs.push({
          jsonrpc: "2.0",
          id: blockNumberStrId,
          method: "eth_getBlockByNumber",
          params: [blockNumberStrParam, isTransactionsPresent],
        });
      }

      let results: (undefined | ResultBlocks)[] = [];
      try {
        results = await provider.post<typeof results>(reqs, {
          timeout: chainCfg.timeout,
          retries: chainCfg.retries,
        });
      } catch (e: HttpClientError | any) {
        provider.setProviderOffline();
        throw e;
      }

      for (let result of results) {
        // If result is not present or error is present, we throw an error to re-try get the transaction
        if (!result || !result.result || result.error) {
          const requestDetails = JSON.stringify(reqs.find((r) => r.id === result?.id));
          this.logger.error(
            `[${chain}][getBlocks] Cannot process this tx: ${requestDetails}, error ${JSON.stringify(
              result?.error
            )} on ${provider.getUrl()}`
          );
          provider.setProviderOffline();
          throw new Error("Unable to parse result of eth_getBlockByNumber");
        }
        combinedResults.push(result);
      }
    }

    if (combinedResults && combinedResults.length) {
      return combinedResults
        .map(
          (
            response: undefined | { id: string; result?: EvmBlock; error?: ErrorBlock },
            idx: number
          ) => {
            // Karura is getting 6969 errors for some blocks, so we'll just return empty blocks for those instead of throwing an error.
            // We take the timestamp from the previous block, which is not ideal but should be fine.
            if (response?.error && response.error?.code && response.error.code === 6969) {
              return {
                hash: "",
                number: BigInt(response.id),
                timestamp: Date.now(),
              };
            }
            if (
              response?.result &&
              response.result?.hash &&
              response.result.number &&
              response.result.timestamp
            ) {
              return {
                hash: response.result.hash,
                number: BigInt(response.result.number),
                timestamp: Number(response.result.timestamp),
                transactions: response.result.transactions,
              };
            }

            this.logger.error(
              `[${chain}][getBlocks] Got error ${
                response?.error?.message
              } for eth_getBlockByNumber for ${response?.id ?? idx} on ${provider.getUrl()}`
            );
            provider.setProviderOffline();
            throw new Error("Unable to parse result of eth_getBlockByNumber");
          }
        )
        .reduce((acc: Record<string, EvmBlock>, block: EvmBlock) => {
          acc[block.hash] = block;
          return acc;
        }, {});
    }

    this.logger.error(
      `[${chain}][getBlocks] Unable to parse ${
        combinedResults?.length ?? 0
      } blocks for eth_getBlockByNumber for numbers ${blockNumbers} on ${provider.getUrl()}`
    );
    provider.setProviderOffline();
    throw new Error("Unable to parse result of eth_getBlockByNumber");
  }

  async getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]> {
    let parsedFilters: ParsedFilters = {
      topics: filter.topics,
      fromBlock: `${HEXADECIMAL_PREFIX}${filter.fromBlock.toString(16)}`,
      toBlock: `${HEXADECIMAL_PREFIX}${filter.toBlock.toString(16)}`,
    };

    if (filter.addresses.length > 0) {
      parsedFilters.address = filter.addresses;
    }

    const provider = getChainProvider(chain, this.pool);
    const chainCfg = this.getCurrentChain(chain);
    let response: { result: Log[]; error?: ErrorBlock };

    try {
      response = await provider.post<typeof response>(
        {
          jsonrpc: "2.0",
          method: "eth_getLogs",
          params: [parsedFilters],
          id: 1,
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      provider.setProviderOffline();
      throw e;
    }

    if (!response || !response.result || response.error) {
      this.logger.error(
        `[${chain}][getFilteredLogs] Error fetching logs with message: ${
          response?.error?.message
        }. Filter: ${JSON.stringify(filter)} on ${provider.getUrl()}`
      );
      provider.setProviderOffline();
      throw new Error("Unable to parse result of eth_getLogs");
    }

    const logs = response?.result;
    if (!logs || logs.length === 0) {
      return [];
    }

    this.logger.info(
      `[${chain}][getFilteredLogs] Got ${logs.length} logs for ${this.describeFilter(
        filter
      )} from ${provider.getUrl()}`
    );

    return logs.map((log) => ({
      ...log,
      blockNumber: BigInt(log.blockNumber),
      transactionIndex: log.transactionIndex.toString(),
      chainId: chainCfg.chainId,
      chain,
    }));
  }

  /**
   * Loosely based on the wormhole-dashboard implementation (minus some specially crafted blocks when null result is obtained)
   */
  async getBlock(
    chain: string,
    blockNumberOrTag: EvmTag | bigint,
    isTransactionsPresent: boolean = false
  ): Promise<EvmBlock> {
    const provider = getChainProvider(chain, this.pool);
    const chainCfg = this.getCurrentChain(chain);

    const response = await this.getBlockByNumber(
      provider,
      blockNumberOrTag,
      isTransactionsPresent,
      chainCfg.timeout,
      chainCfg.retries
    );

    const result = response?.result;
    if (result && result.hash && result.number && result.timestamp) {
      // Convert to our domain compatible type
      return {
        number: BigInt(result.number),
        timestamp: Number(result.timestamp),
        hash: result.hash,
        transactions: result.transactions,
      };
    }

    this.logger.error(
      `[${chain}][getBlock] Unable to parse result of eth_getBlockByNumber for ${blockNumberOrTag} on ${provider.getUrl()}. Response error: ${JSON.stringify(
        response
      )}`
    );
    provider.setProviderOffline();
    throw new Error("Unable to parse result of eth_getBlockByNumber");
  }

  /**
   * Get the transaction ReceiptTransaction. Hash param refers to transaction hash
   */
  async getTransactionReceipt(
    chain: string,
    hashNumbers: Set<string>
  ): Promise<Record<string, ReceiptTransaction>> {
    const chainCfg = this.getCurrentChain(chain);
    let results: ResultTransactionReceipt[] = [];
    let id = 1;

    /**
     * This method divide in batches the object to send, because we have one restriction about how many object send to the endpoint
     * the maximum is 10 object per request
     */
    const provider = getChainProvider(chain, this.pool);
    const batches = divideIntoBatches(hashNumbers, TX_BATCH_SIZE);
    let combinedResults: ResultTransactionReceipt[] = [];

    for (const batch of batches) {
      const reqs: any[] = [];
      for (let hash of batch) {
        reqs.push({
          jsonrpc: "2.0",
          id,
          method: "eth_getTransactionReceipt",
          params: [hash],
        });
        id++;
      }

      try {
        results = await provider.post<typeof results>(reqs, {
          timeout: chainCfg.timeout,
          retries: chainCfg.retries,
        });
      } catch (e: HttpClientError | any) {
        throw e;
      }

      for (let result of results) {
        // If result is not present or error is present, we throw an error to re-try get the transaction
        if (!result || !result.result || result.error) {
          const requestDetails = JSON.stringify(reqs.find((r) => r.id === result?.id));
          this.logger.error(
            `[${chain}][getTransactionReceipt] Cannot process this tx: ${requestDetails}, error ${JSON.stringify(
              result.error
            )} on ${provider.getUrl()}`
          );
          provider.setProviderOffline();
          throw new Error("Unable to parse result of eth_getTransactionReceipt");
        }
        combinedResults.push(result);
      }
    }

    if (combinedResults && combinedResults.length) {
      return combinedResults
        .map((response) => {
          if (response.result?.status && response.result?.transactionHash) {
            return {
              effectiveGasPrice: response.result.effectiveGasPrice,
              transactionHash: response.result.transactionHash,
              gasUsed: response.result.gasUsed,
              status: response.result.status,
              logs: response.result.logs,
            };
          }

          this.logger.error(
            `[${chain}][getTransactionReceipt] Got error ${
              response?.error ?? JSON.stringify(response)
            } for eth_getTransactionReceipt for ${JSON.stringify(
              hashNumbers
            )} on ${provider.getUrl()}`
          );
          provider.setProviderOffline();
          throw new Error("Unable to parse result of eth_getTransactionReceipt");
        })
        .reduce(
          (acc: Record<string, ReceiptTransaction>, receiptTransaction: ReceiptTransaction) => {
            acc[receiptTransaction.transactionHash] = receiptTransaction;
            return acc;
          },
          {}
        );
    }
    this.logger.error(
      `[${chain}][getTransactionReceipt] Unable to parse result of eth_getTransactionReceipt 
      for ${JSON.stringify(hashNumbers)} on ${provider.getUrl()}`
    );
    provider.setProviderOffline();
    throw new Error("Unable to parse result of eth_getTransactionReceipt");
  }

  protected handleError(chain: string, e: any, method: string, apiMethod: string) {
    const chainCfg = this.getCurrentChain(chain);
    if (e instanceof HttpClientError) {
      this.logger.error(
        `[${chain}][${method}] Got ${e.status} from ${chainCfg.rpc.hostname}/${apiMethod}. ${
          e?.message ?? `${e?.message}`
        }`
      );
    } else {
      this.logger.error(
        `[${chain}][${method}] Got error ${e} from ${chainCfg.rpc.hostname}/${apiMethod}`
      );
    }
  }

  protected getCurrentChain(chain: string) {
    const cfg = this.cfg.chains[chain];
    return {
      chainId: cfg.chainId,
      rpc: new URL(cfg.rpcs[0]),
      timeout: cfg.timeout ?? 10_000,
      retries: cfg.retries ?? 2,
    };
  }

  // Private method to not duplicate code getting block height value
  private async getBlockByNumber(
    provider: InstrumentedHttpProvider,
    blockNumberOrTag: EvmTag | bigint,
    isTransactionsPresent: boolean = false,
    timeout: number = 10_000,
    retries: number = 2
  ) {
    let response: { result?: EvmBlock; error?: ErrorBlock };
    const blockNumberParam =
      typeof blockNumberOrTag === "bigint"
        ? `${HEXADECIMAL_PREFIX}${blockNumberOrTag.toString(16)}`
        : blockNumberOrTag;

    try {
      return await provider.post<typeof response>(
        {
          jsonrpc: "2.0",
          method: "eth_getBlockByNumber",
          params: [blockNumberParam, isTransactionsPresent], // this means we'll get a light block (no txs)
          id: 1,
        },
        { timeout: timeout, retries: retries }
      );
    } catch (e: HttpClientError | any) {
      provider.setProviderOffline();
      throw e;
    }
  }

  private describeFilter(filter: EvmLogFilter): string {
    return `[addresses:${filter.addresses}][topics:${filter.topics}][blocks:${filter.fromBlock} - ${filter.toBlock}]`;
  }
}

type ErrorBlock = {
  code: number; //6969,
  message: string; //'Error: No response received from RPC endpoint in 60s'
};

type Log = {
  blockNumber: string;
  blockHash: string;
  transactionIndex: number;
  removed: boolean;
  address: string;
  data: string;
  topics: Array<string>;
  transactionHash: string;
  logIndex: number;
};

type ResultTransactionReceipt = {
  id: string;
  result: ReceiptTransaction;
  error?: ErrorBlock;
};

type ResultBlocks = {
  id: string;
  result?: EvmBlock;
  error?: ErrorBlock;
};

type ParsedFilters = {
  fromBlock: string;
  toBlock: string;
  address?: string[];
  topics: string[];
};
