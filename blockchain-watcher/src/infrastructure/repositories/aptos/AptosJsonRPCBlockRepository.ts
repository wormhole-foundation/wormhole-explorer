import { Range, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { AptosRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";
import {
  AptosTransactionByVersion,
  AptosBlockByVersion,
  AptosTransaction,
  AptosEvent,
} from "../../../domain/entities/aptos";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

export class AptosJsonRPCBlockRepository implements AptosRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "AptosJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getEventsByEventHandle(
    range: Range | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    try {
      let results: AptosEvent[] = [];

      const from = range?.from ? Number(range?.from) : undefined;
      const limit = range?.limit ? Number(range?.limit) : undefined;

      let endpoint = `/accounts/${filter.address}/events/${filter.event}/${filter.fieldName}`;

      if (from) {
        endpoint = `${endpoint}?start=${from}`;
      }

      if (from && limit) {
        endpoint = `${endpoint}&limit=${limit}`;
      }

      results = await this.pool.get().get<typeof results>({ endpoint });
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
    records: AptosEvent[] | AptosTransaction[],
    filter: TransactionFilter
  ): Promise<AptosTransaction[]> {
    try {
      const transactions = await Promise.all(
        records.map(async (event) => {
          const txEndpoint = `/transactions/by_version/${Number(event.version)}`;
          const blockEndpoint = `/blocks/by_version/${Number(event.version)}`;

          let txResult: AptosTransactionByVersion = {};
          let blockResult: AptosBlockByVersion = {};

          txResult = await this.pool.get().get<typeof txResult>({ endpoint: txEndpoint });
          blockResult = await this.pool.get().get<typeof blockResult>({ endpoint: blockEndpoint });

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
      const from = range?.from ? Number(range?.from) : undefined;
      const limit = range?.limit ? Number(range?.limit) : undefined;

      let results: AptosTransaction[] = [];

      let endpoint = `/transactions`;

      if (from) {
        endpoint = `${endpoint}?start=${from}`;
      }

      if (from && limit) {
        endpoint = `${endpoint}&limit=${limit}`;
      }

      results = await this.pool.get().get<typeof results>({ endpoint });
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
