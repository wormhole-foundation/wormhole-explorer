import { EvmBlock, EvmLogFilter, EvmLog, EvmTag } from "../../domain/entities";
import { EvmBlockRepository } from "../../domain/repositories";
import { AxiosInstance } from "axios";
import winston from "winston";

const headers = {
  "Content-Type": "application/json",
};

/**
 * EvmJsonRPCBlockRepository is a repository that uses a JSON RPC endpoint to fetch blocks.
 * On the reliability side, only knows how to timeout.
 */
export class EvmJsonRPCBlockRepository implements EvmBlockRepository {
  private axios: AxiosInstance;
  private rpc: URL;
  private timeout: number;
  private readonly logger = winston.child({ module: "EvmJsonRPCBlockRepository" });

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, axios: AxiosInstance) {
    this.axios = axios;
    this.rpc = new URL(cfg.rpc);
    this.timeout = cfg.timeout ?? 10_000;
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
      const blockNumberStr = blockNumber.toString();
      reqs.push({
        jsonrpc: "2.0",
        id: blockNumberStr,
        method: "eth_getBlockByNumber",
        params: [blockNumberStr, false],
      });
    }
    const response = await this.axios.post(this.rpc.href, reqs, this.getRequestOptions());
    if (response.status !== 200) {
      this.logger.error(
        `Got ${response.status} from ${this.rpc.hostname}/eth_getBlockByNumber. ${
          response?.data?.error?.message ?? `${response?.data?.error.message}`
        }`
      );
    }

    const results = response?.data;
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
      fromBlock: `0x${filter.fromBlock.toString(16)}`,
      toBlock: `0x${filter.toBlock.toString(16)}`,
    };

    let response = await this.axios.post<{ result: Log[]; error?: ErrorBlock }>(
      this.rpc.href,
      {
        jsonrpc: "2.0",
        method: "eth_getLogs",
        params: [parsedFilters],
        id: 1,
      },
      this.getRequestOptions()
    );

    if (response.status !== 200 || response?.data.error) {
      const msg = `Got error ${response?.data?.error?.message} for ${this.describeFilter(
        filter
      )} from ${this.rpc.hostname}/eth_getLogs`;
      this.logger.error(`Got ${response.status} from ${this.rpc.hostname}. ${msg}`);

      throw new Error(msg);
    }
    const logs = response?.data?.result;
    this.logger.info(
      `Got ${logs?.length} logs for ${this.describeFilter(filter)} from ${this.rpc.hostname}`
    );

    return logs.map((log) => ({
      ...log,
      blockNumber: BigInt(log.blockNumber),
      transactionIndex: log.transactionIndex.toString(),
    }));
  }

  private describeFilter(filter: EvmLogFilter): string {
    return `[addresses:${filter.addresses}][topics:${filter.topics}][blocks:${filter.fromBlock} - ${filter.toBlock}]`;
  }

  /**
   * Loosely based on the wormhole-dashboard implementation (minus some specially crafted blocks when null result is obtained)
   */
  private async getBlock(blockNumberOrTag: bigint | EvmTag): Promise<EvmBlock> {
    let response = await this.axios.post(
      this.rpc.href,
      {
        jsonrpc: "2.0",
        method: "eth_getBlockByNumber",
        params: [blockNumberOrTag.toString(), false], // this means we'll get a light block (no txs)
        id: 1,
      },
      this.getRequestOptions()
    );

    if (response.status !== 200 || response?.data?.error) {
      this.logger.error(
        `Got ${response.status} from ${this.rpc.hostname}/eth_getBlockByNumber. ${
          response?.data?.error?.message ?? `${response?.data?.error.message}`
        }`
      );
    }

    const result = response?.data?.result;

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

  private getRequestOptions() {
    return { headers, timeout: this.timeout, signal: AbortSignal.timeout(this.timeout) };
  }
}

export type EvmJsonRPCBlockRepositoryCfg = {
  rpc: string;
  timeout?: number;
  chain: string;
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
