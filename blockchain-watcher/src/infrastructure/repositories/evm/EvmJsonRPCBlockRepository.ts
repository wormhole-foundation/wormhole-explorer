import { EvmBlock, EvmLogFilter, EvmLog, EvmTag } from "../../../domain/entities";
import { EvmBlockRepository } from "../../../domain/repositories";
import winston from "../../log";
import { HttpClient } from "../../rpc/http/HttpClient";
import { HttpClientError } from "../../errors/HttpClientError";
import { ChainRPCConfig } from "../../config";

/**
 * EvmJsonRPCBlockRepository is a repository that uses a JSON RPC endpoint to fetch blocks.
 * On the reliability side, only knows how to timeout.
 */

const HEXADECIMAL_PREFIX = "0x";

export class EvmJsonRPCBlockRepository implements EvmBlockRepository {
  protected httpClient: HttpClient;
  private cfg: EvmJsonRPCBlockRepositoryCfg;
  protected readonly logger;

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    this.httpClient = httpClient;
    this.cfg = cfg;

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
  async getBlocks(chain: string, blockNumbers: Set<bigint>): Promise<Record<string, EvmBlock>> {
    if (!blockNumbers.size) return {};

    const reqs: any[] = [];
    for (let blockNumber of blockNumbers) {
      const blockNumberStrParam = `${HEXADECIMAL_PREFIX}${blockNumber.toString(16)}`;
      const blockNumberStrId = blockNumber.toString();

      reqs.push({
        jsonrpc: "2.0",
        id: blockNumberStrId,
        method: "eth_getBlockByNumber",
        params: [blockNumberStrParam, false],
      });
    }

    const chainCfg = this.getCurrentChain(chain);
    let results: (undefined | { id: string; result?: EvmBlock; error?: ErrorBlock })[];
    try {
      results = await this.httpClient.post<typeof results>(chainCfg.rpc.href, reqs, {
        timeout: chainCfg.timeout,
        retries: chainCfg.retries,
      });
    } catch (e: HttpClientError | any) {
      this.handleError(chain, e, "getBlocks", "eth_getBlockByNumber");
      throw e;
    }

    if (results && results.length) {
      return results
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
              };
            }

            const msg = `[${chain}][getBlocks] Got error ${
              response?.error?.message
            } for eth_getBlockByNumber for ${response?.id ?? reqs[idx].id} on ${
              chainCfg.rpc.hostname
            }`;

            this.logger.error(msg);

            throw new Error(
              `Unable to parse result of eth_getBlockByNumber[${chain}] for ${
                response?.id ?? reqs[idx].id
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
        results?.length ?? 0
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
      response = await this.httpClient.post<typeof response>(
        chainCfg.rpc.href,
        {
          jsonrpc: "2.0",
          method: "eth_getLogs",
          params: [parsedFilters],
          id: 1,
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      this.handleError(chain, e, "getFilteredLogs", "eth_getLogs");
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
  private async getBlock(chain: string, blockNumberOrTag: EvmTag): Promise<EvmBlock> {
    const chainCfg = this.getCurrentChain(chain);
    let response: { result?: EvmBlock; error?: ErrorBlock };
    try {
      response = await this.httpClient.post<typeof response>(
        chainCfg.rpc.href,
        {
          jsonrpc: "2.0",
          method: "eth_getBlockByNumber",
          params: [blockNumberOrTag, false], // this means we'll get a light block (no txs)
          id: 1,
        },
        { timeout: chainCfg.timeout, retries: chainCfg.retries }
      );
    } catch (e: HttpClientError | any) {
      this.handleError(chain, e, "getBlock", "eth_getBlockByNumber");
      throw e;
    }

    const result = response?.result;

    if (result && result.hash && result.number && result.timestamp) {
      // Convert to our domain compatible type
      return {
        number: BigInt(result.number),
        timestamp: Number(result.timestamp),
        hash: result.hash,
      };
    }
    throw new Error(
      `Unable to parse result of eth_getBlockByNumber for ${blockNumberOrTag} on ${chainCfg.rpc}`
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
