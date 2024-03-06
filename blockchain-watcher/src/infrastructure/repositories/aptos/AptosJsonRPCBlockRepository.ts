import { Sequence, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { InstrumentedAptosProvider } from "../../rpc/http/InstrumentedAptosProvider";
import { coalesceChainId } from "@certusone/wormhole-sdk/lib/cjs/utils/consts";
import { AptosEvent } from "../../../domain/entities/aptos";
import winston from "winston";

export class AptosJsonRPCBlockRepository {
  private readonly logger: winston.Logger;

  constructor(private readonly client: InstrumentedAptosProvider) {
    this.logger = winston.child({ module: "AptosJsonRPCBlockRepository" });
  }

  async getSequenceNumber(
    range: Sequence | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    try {
      const fromSequence = range?.fromSequence ? Number(range?.fromSequence) : undefined;
      const toSequence = range?.toSequence ? Number(range?.toSequence) : undefined;

      const results = await this.client.getEventsByEventHandle(
        filter.address,
        filter.event,
        filter.fieldName,
        fromSequence,
        toSequence
      );
      return results;
    } catch (e) {
      this.handleError(e, "getSequenceNumber");
      throw e;
    }
  }

  async getTransactionsForVersions(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<TransactionsByVersion[]> {
    try {
      const transactionsByVersion: TransactionsByVersion[] = [];

      for (const event of events) {
        const transaction = await this.client.getTransactionByVersion(Number(event.version));
        const block = await this.client.getBlockByVersion(Number(event.version));

        const wormholeEvent = transaction.events.find((tx: any) => tx.type === filter.type);

        const tx = {
          consistencyLevel: event.data.consistency_level,
          blockHeight: block.block_height,
          timestamp: wormholeEvent.data.timestamp,
          blockTime: wormholeEvent.data.timestamp,
          sequence: wormholeEvent.data.sequence,
          version: transaction.version,
          payload: wormholeEvent.data.payload,
          address: filter.address,
          sender: wormholeEvent.data.sender,
          status: transaction.success,
          events: transaction.events,
          nonce: wormholeEvent.data.nonce,
          hash: transaction.hash,
        };
        transactionsByVersion.push(tx);
      }

      return transactionsByVersion;
    } catch (e) {
      this.handleError(e, "getTransactionsForVersions");
      throw e;
    }
  }

  async getTransactions(limit: number): Promise<any[]> {
    try {
      const results = await this.client.getTransactions(limit);
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

export type TransactionsByVersion = {
  consistencyLevel: number;
  blockHeight: bigint;
  timestamp: number;
  blockTime: number;
  sequence: string;
  version: string;
  payload: string;
  address: string;
  sender: string;
  status?: boolean;
  events: any;
  nonce: string;
  hash: string;
};

// TODO: Remove
const makeVaaKey = (transactionHash: string, emitter: string, seq: string): string =>
  `${transactionHash}:${coalesceChainId("aptos")}/${emitter}/${seq}`;
