import {
  AptosEvent,
  AptosTransaction,
  AptosTransactionByRange,
} from "../../../domain/entities/aptos";
import { InstrumentedAptosProvider } from "../../rpc/http/InstrumentedAptosProvider";
import { Range, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { AptosRepository } from "../../../domain/repositories";
import winston from "winston";

export class AptosJsonRPCBlockRepository implements AptosRepository {
  private readonly logger: winston.Logger;

  constructor(private readonly client: InstrumentedAptosProvider) {
    this.logger = winston.child({ module: "AptosJsonRPCBlockRepository" });
  }

  async getEventsByEventHandle(
    range: Range | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    try {
      const from = range?.from ? Number(range?.from) : undefined;
      const limit = range?.limit ? Number(range?.limit) : undefined;

      const results = await this.client.getEventsByEventHandle(
        filter.address,
        filter.event!,
        filter.fieldName,
        from,
        limit
      );

      return results;
    } catch (e) {
      this.handleError(e, "getEventsByEventHandle");
      throw e;
    }
  }

  async getTransactionsByVersion(
    events: AptosEvent[] | AptosTransactionByRange[],
    filter: TransactionFilter
  ): Promise<AptosTransaction[]> {
    try {
      const transactions = await Promise.all(
        events.map(async (event) => {
          const transaction = await this.client.getTransactionByVersion(Number(event.version));
          const block = await this.client.getBlockByVersion(Number(event.version));

          return {
            consistencyLevel: event?.data?.consistency_level,
            blockHeight: BigInt(block.block_height),
            version: transaction.version!,
            address: event.events ? event.events[0].guid.account_address : filter.address,
            status: transaction.success,
            events: transaction.events,
            hash: transaction.hash,
            type: filter.type,
            payload: transaction.payload,
            to: filter.address,
          };
        })
      );

      return transactions;
    } catch (e) {
      this.handleError(e, "getTransactionsByVersion");
      throw e;
    }
  }

  async getTransactions(range: Range): Promise<AptosTransactionByRange[]> {
    try {
      const results = await this.client.getTransactions(range);
      return results;
    } catch (e) {
      this.handleError(e, "getTransactions");
      throw e;
    }
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[aptos] Error calling ${method}: ${e.message}`);
  }
}
