import { Range, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { ProviderPoolDecorator } from "../../rpc/http/ProviderPoolDecorator";
import { AptosRepository } from "../../../domain/repositories";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import winston from "winston";
import {
  AptosTransactionByVersion,
  AptosBlockByVersion,
  AptosTransaction,
  AptosEvent,
} from "../../../domain/entities/aptos";

let TRANSACTION_BY_VERSION_ENDPOINT = "/transactions/by_version";
let BLOCK_BY_VERSION_ENDPOINT = "/blocks/by_version";
let TRANSACTION_ENDPOINT = "/transactions";
let ACCOUNT_ENDPOINT = "/accounts";

type ProviderPoolMap = ProviderPoolDecorator<InstrumentedHttpProvider>;

export class AptosJsonRPCBlockRepository implements AptosRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPoolDecorator<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "AptosJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async healthCheck(
    chain: string,
    finality: string,
    cursor: bigint
  ): Promise<ProviderHealthCheck[]> {
    // If the cursor is not set yet, we try again later
    if (!cursor) {
      return [];
    }
    const blockEndpoint = `${BLOCK_BY_VERSION_ENDPOINT}/${Number(cursor)}`;
    const providers = this.pool.getProviders();
    let blockResult: AptosBlockByVersion = {};
    const providersHealthCheck: ProviderHealthCheck[] = [];

    for (const provider of providers) {
      try {
        const requestStartTime = performance.now();
        blockResult = await this.pool.get().get<typeof blockResult>(blockEndpoint);
        const requestEndTime = performance.now();

        const height = blockResult.block_height ? BigInt(blockResult.block_height) : undefined;

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
    this.pool.setProviders(chain, providers, providersHealthCheck, cursor);
    return providersHealthCheck;
  }

  async getEventsByEventHandle(
    range: Range | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    try {
      let endpoint = `${ACCOUNT_ENDPOINT}/${filter.address}/events/${filter.event}/${filter.fieldName}`;
      let results: AptosEvent[] = [];

      results = await this.pool.get().get<typeof results>(endpoint, {
        limit: range?.limit,
        start: range?.from,
      });
      return results;
    } catch (e) {
      this.handleError(
        `Range params: ${JSON.stringify(range)}, error: ${e}`,
        "getEventsByEventHandle"
      );
      throw e;
    }
  }

  async getTransactionsByVersion(
    records: AptosEvent[] | AptosTransaction[]
  ): Promise<AptosTransaction[]> {
    try {
      const transactions = await Promise.all(
        records.map(async (event) => {
          const txEndpoint = `${TRANSACTION_BY_VERSION_ENDPOINT}/${Number(event.version)}`;
          const blockEndpoint = `${BLOCK_BY_VERSION_ENDPOINT}/${Number(event.version)}`;

          let txResult: AptosTransactionByVersion = {};
          let blockResult: AptosBlockByVersion = {};

          txResult = await this.pool.get().get<typeof txResult>(txEndpoint);
          blockResult = await this.pool.get().get<typeof blockResult>(blockEndpoint);

          return {
            blockHeight: BigInt(blockResult.block_height!),
            version: txResult.version!,
            status: txResult.success,
            events: txResult.events,
            hash: txResult.hash!,
            payload: txResult.payload,
          };
        })
      );

      return transactions;
    } catch (e) {
      this.handleError(e, "getTransactionsByVersion");
      throw e;
    }
  }

  async getTransactions(range: Range): Promise<AptosTransaction[]> {
    try {
      let results: AptosTransaction[] = [];

      results = await this.pool.get().get<typeof results>(TRANSACTION_ENDPOINT, {
        limit: range?.limit,
        start: range?.from,
      });
      return results;
    } catch (e) {
      this.handleError(`Range params: ${JSON.stringify(range)}, error: ${e}`, "getTransactions");
      throw e;
    }
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[aptos] Error calling ${method}: ${e.message ?? e}`);
  }
}
