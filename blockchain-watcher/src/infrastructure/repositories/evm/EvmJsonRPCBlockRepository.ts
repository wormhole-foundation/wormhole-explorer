import {
  EvmBlock,
  EvmLogFilter,
  EvmLog,
  EvmTag,
  ReceiptTransaction,
} from "../../../domain/entities";
import { EvmBlockRepository } from "../../../domain/repositories";
import winston from "../../log";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { HttpClientError } from "../../errors/HttpClientError";
import { ChainRPCConfig } from "../../config";
import { divideIntoBatches } from "../common/utils";
import { ProviderPool } from "@xlabs/rpc-pool";

/**
 * EvmJsonRPCBlockRepository is a repository that uses a JSON RPC endpoint to fetch blocks.
 * On the reliability side, only knows how to timeout.
 */

const HEXADECIMAL_PREFIX = "0x";
const TX_BATCH_SIZE = 10;

export type ProviderPoolMap = Record<string, ProviderPool<InstrumentedHttpProvider>>;

export class EvmJsonRPCBlockRepository implements EvmBlockRepository {
  protected pool: ProviderPoolMap;
  protected cfg: EvmJsonRPCBlockRepositoryCfg;
  protected readonly logger;

  constructor(
    cfg: EvmJsonRPCBlockRepositoryCfg,
    pool: Record<string, ProviderPool<InstrumentedHttpProvider>>
  ) {
    this.cfg = cfg;
    this.pool = pool;

    this.logger = winston.child({ module: "EvmJsonRPCBlockRepository" });
    this.logger.info(`Created for ${Object.keys(this.cfg.chains)}`);
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
        results = await this.getChainProvider(chain).post<typeof results>(reqs, {
          timeout: chainCfg.timeout,
          retries: chainCfg.retries,
        });
      } catch (e: HttpClientError | any) {
        throw e;
      }

      for (let result of results) {
        if (result) {
          combinedResults.push(result);
        }
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
            if (
              (response && response.result === null) ||
              (response?.error && response.error?.code && response.error.code === 6969)
            ) {
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

            const msg = `[${chain}][getBlocks] Got error ${
              response?.error?.message
            } for eth_getBlockByNumber for ${response?.id ?? idx} on ${chainCfg.rpc.hostname}`;

            this.logger.error(msg);

            throw new Error(
              `Unable to parse result of eth_getBlockByNumber[${chain}] for ${
                response?.id ?? idx
              }: ${msg}`
            );
          }
        )
        .reduce((acc: Record<string, EvmBlock>, block: EvmBlock) => {
          acc[block.hash] = block;
          return acc;
        }, {});
    }

    throw new Error(
      `Unable to parse ${
        combinedResults?.length ?? 0
      } blocks for eth_getBlockByNumber for numbers ${blockNumbers} on ${chainCfg.rpc.hostname}`
    );
  }

  async getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]> {
    const parsedFilters = {
      topics: filter.topics,
      address: filter.addresses,
      fromBlock: `${HEXADECIMAL_PREFIX}${filter.fromBlock.toString(16)}`,
      toBlock: `${HEXADECIMAL_PREFIX}${filter.toBlock.toString(16)}`,
    };

    const chainCfg = this.getCurrentChain(chain);
    let response: { result: Log[]; error?: ErrorBlock };
    try {
      response = await this.getChainProvider(chain).post<typeof response>(
        {
          jsonrpc: "2.0",
          method: "eth_getLogs",
          params: [parsedFilters],
          id: 1,
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      throw e;
    }

    const logs = response?.result;
    this.logger.info(
      `[${chain}][getFilteredLogs] Got ${logs?.length} logs for ${this.describeFilter(
        filter
      )} from ${chainCfg.rpc.hostname}`
    );

    return logs
      ? logs.map((log) => ({
          ...log,
          blockNumber: BigInt(log.blockNumber),
          transactionIndex: log.transactionIndex.toString(),
          chainId: chainCfg.chainId,
        }))
      : [];
  }

  private describeFilter(filter: EvmLogFilter): string {
    return `[addresses:${filter.addresses}][topics:${filter.topics}][blocks:${filter.fromBlock} - ${filter.toBlock}]`;
  }

  /**
   * Loosely based on the wormhole-dashboard implementation (minus some specially crafted blocks when null result is obtained)
   */
  async getBlock(
    chain: string,
    blockNumberOrTag: EvmTag | bigint,
    isTransactionsPresent: boolean = false
  ): Promise<EvmBlock> {
    const blockNumberParam =
      typeof blockNumberOrTag === "bigint"
        ? `${HEXADECIMAL_PREFIX}${blockNumberOrTag.toString(16)}`
        : blockNumberOrTag;

    const chainCfg = this.getCurrentChain(chain);
    let response: { result?: EvmBlock; error?: ErrorBlock };
    try {
      response = await this.getChainProvider(chain).post<typeof response>(
        {
          jsonrpc: "2.0",
          method: "eth_getBlockByNumber",
          params: [blockNumberParam, isTransactionsPresent], // this means we'll get a light block (no txs)
          id: 1,
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      throw e;
    }

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
    throw new Error(
      `Unable to parse result of eth_getBlockByNumber for ${blockNumberOrTag} on ${
        chainCfg.rpc
      }. Response error: ${JSON.stringify(response)}`
    );
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
        results = await this.getChainProvider(chain).post<typeof results>(reqs, {
          timeout: chainCfg.timeout,
          retries: chainCfg.retries,
        });
      } catch (e: HttpClientError | any) {
        throw e;
      }

      for (let result of results) {
        if (result) {
          combinedResults.push(result);
        }
      }
    }

    if (combinedResults && combinedResults.length) {
      return combinedResults
        .map((response) => {
          if (response.result?.status && response.result?.transactionHash) {
            return {
              status: response.result.status,
              transactionHash: response.result.transactionHash,
              logs: response.result.logs,
            };
          }

          const msg = `[${chain}][getTransactionReceipt] Got error ${
            response?.error
          } for eth_getTransactionReceipt for ${JSON.stringify(hashNumbers)} on ${
            chainCfg.rpc.hostname
          }`;

          this.logger.error(msg);

          throw new Error(
            `Unable to parse result of eth_getTransactionReceipt[${chain}] for ${response?.result}: ${msg}`
          );
        })
        .reduce(
          (acc: Record<string, ReceiptTransaction>, receiptTransaction: ReceiptTransaction) => {
            acc[receiptTransaction.transactionHash] = receiptTransaction;
            return acc;
          },
          {}
        );
    }
    throw new Error(
      `Unable to parse result of eth_getTransactionReceipt for ${JSON.stringify(hashNumbers)} on ${
        chainCfg.rpc
      }. Result error: ${JSON.stringify(combinedResults)}`
    );
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

  protected getChainProvider(chain: string): InstrumentedHttpProvider {
    const pool = this.pool[chain];
    if (!pool) {
      throw new Error(`No provider pool configured for chain ${chain}`);
    }
    return pool.get();
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
}

export type EvmJsonRPCBlockRepositoryCfg = {
  chains: Record<string, ChainRPCConfig>;
};

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
  result: ReceiptTransaction;
  error?: ErrorBlock;
};

type ResultBlocks = {
  id: string;
  result?: EvmBlock;
  error?: ErrorBlock;
};
