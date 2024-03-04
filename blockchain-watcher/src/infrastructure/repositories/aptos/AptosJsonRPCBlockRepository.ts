import { Sequence, TransactionFilter } from "../../../domain/actions/aptos/PollAptosTransactions";
import { InstrumentedAptosProvider } from "../../rpc/http/InstrumentedAptosProvider";
import { coalesceChainId } from "@certusone/wormhole-sdk/lib/cjs/utils/consts";
import { AptosEvent } from "../../../domain/entities/aptos";
import winston from "winston";

export class AptosJsonRPCBlockRepository {
  private readonly logger: winston.Logger;

  constructor(private readonly client: InstrumentedAptosProvider) {
    this.logger = winston.child({ module: "AptossonRPCBlockRepository" });
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
    const transactionsByVersion: TransactionsByVersion[] = [];

    for (const event of events) {
      const transaction = await this.client.getTransactionByVersion(Number(event.version));
      const block = await this.client.getBlockByVersion(Number(event.version));

      const tx = {
        consistencyLevel: event.data.consistency_level,
        blockHeight: block.block_height,
        timestamp: transaction.timestamp,
        blockTime: block.block_timestamp,
        sequence: transaction.sequence_number,
        version: transaction.version,
        payload: transaction.payload,
        address: filter.address,
        sender: transaction.sender,
        status: transaction.success,
        events: transaction.events,
        nonce: event.data.nonce,
        hash: transaction.hash,
      };
      transactionsByVersion.push(tx);
    }

    return transactionsByVersion;
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[aptos] Error calling ${method}: ${e.message}`);
  }
}

export type TransactionsByVersion = {
  consistencyLevel?: number;
  blockHeight?: bigint;
  timestamp?: string;
  blockTime: number;
  sequence?: string;
  version?: string;
  payload?: string;
  address?: string;
  sender?: string;
  status?: boolean;
  events?: any;
  nonce?: string;
  hash?: string;
};

const makeVaaKey = (transactionHash: string, emitter: string, seq: string): string =>
  `${transactionHash}:${coalesceChainId("aptos")}/${emitter}/${seq}`;
