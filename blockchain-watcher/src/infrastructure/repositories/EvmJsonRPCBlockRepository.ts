import { EvmBlock, EvmLogFilter, EvmLog, EvmTag } from "../../domain/entities";
import { EvmBlockRepository } from "../../domain/repositories";
import winston from "../log";
import { HttpClient, HttpClientError } from "../http/HttpClient";

/**
 * EvmJsonRPCBlockRepository is a repository that uses a JSON RPC endpoint to fetch blocks.
 * On the reliability side, only knows how to timeout.
 */

const HEXADECIMAL_PREFIX = "0x";

export class EvmJsonRPCBlockRepository implements EvmBlockRepository {
  private httpClient: HttpClient;
  private rpc: URL;
  private readonly logger = winston.child({ module: "EvmJsonRPCBlockRepository" });

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    this.httpClient = httpClient;
    this.rpc = new URL(cfg.rpc);
    this.logger = winston.child({ module: "EvmJsonRPCBlockRepository", chain: cfg.chain });
  }

  async getBlockHeight(finality: EvmTag): Promise<bigint> {
    const block: EvmBlock = await this.getBlock(finality);
    return block.number;
  }

  /**
   * Get blocks by block number.
   * @param blockNumbers
   * @returns a record of block hash -> EvmBlock
   */
  async getBlocks(blockNumbers: Set<bigint>): Promise<Record<string, EvmBlock>> {
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

    let results: (undefined | { id: string; result?: EvmBlock; error?: ErrorBlock })[];
    try {
      results = await this.httpClient.post<typeof results>(this.rpc.href, reqs);
    } catch (e: HttpClientError | any) {
      this.handleError(e, "eth_getBlockByNumber");
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
              this.logger.warn;
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

            const msg = `Got error ${response?.error?.message} for eth_getBlockByNumber for ${
              response?.id ?? reqs[idx].id
            } on ${this.rpc.hostname}`;

            this.logger.error(msg);

            throw new Error(
              `Unable to parse result of eth_getBlockByNumber for ${
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
      } blocks for eth_getBlockByNumber for numbers ${blockNumbers} on ${this.rpc.hostname}`
    );
  }

  async getFilteredLogs(filter: EvmLogFilter): Promise<EvmLog[]> {
    const parsedFilters = {
      topics: filter.topics,
      address: filter.addresses,
      fromBlock: `${HEXADECIMAL_PREFIX}${filter.fromBlock.toString(16)}`,
      toBlock: `${HEXADECIMAL_PREFIX}${filter.toBlock.toString(16)}`,
    };

    let response: { result: Log[]; error?: ErrorBlock };
    try {
      response = await this.httpClient.post<typeof response>(this.rpc.href, {
        jsonrpc: "2.0",
        method: "eth_getLogs",
        params: [parsedFilters],
        id: 1,
      });
    } catch (e: HttpClientError | any) {
      this.handleError(e, "eth_getLogs");
      throw e;
    }

    const logs = response?.result;
    this.logger.info(
      `Got ${logs?.length} logs for ${this.describeFilter(filter)} from ${this.rpc.hostname}`
    );

    return logs.map((log) => ({
      ...log,
      blockNumber: BigInt(log.blockNumber),
      transactionIndex: log.transactionIndex.toString(),
      chainId: 1,
    }));
  }

  private describeFilter(filter: EvmLogFilter): string {
    return `[addresses:${filter.addresses}][topics:${filter.topics}][blocks:${filter.fromBlock} - ${filter.toBlock}]`;
  }

  /**
   * Loosely based on the wormhole-dashboard implementation (minus some specially crafted blocks when null result is obtained)
   */
  private async getBlock(blockNumberOrTag: EvmTag): Promise<EvmBlock> {
    let response: { result?: EvmBlock; error?: ErrorBlock };
    try {
      response = await this.httpClient.post<typeof response>(this.rpc.href, {
        jsonrpc: "2.0",
        method: "eth_getBlockByNumber",
        params: [blockNumberOrTag, false], // this means we'll get a light block (no txs)
        id: 1,
      });
    } catch (e: HttpClientError | any) {
      this.handleError(e, "eth_getBlockByNumber");
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
      `Unable to parse result of eth_getBlockByNumber for ${blockNumberOrTag} on ${this.rpc}`
    );
  }

  private handleError(e: any, method: string) {
    if (e instanceof HttpClientError) {
      this.logger.error(
        `Got ${e.status} from ${this.rpc.hostname}/${method}. ${e?.message ?? `${e?.message}`}`
      );
    } else {
      this.logger.error(`Got error ${e} from ${this.rpc.hostname}/${method}`);
    }
  }
}

export type EvmJsonRPCBlockRepositoryCfg = {
  rpc: string;
  timeout?: number;
  chain: string;
  chainId: number;
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
