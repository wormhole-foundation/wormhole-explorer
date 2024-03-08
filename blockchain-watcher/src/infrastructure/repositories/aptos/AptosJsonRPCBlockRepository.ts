import { InstrumentedAptosProvider } from "../../rpc/http/InstrumentedAptosProvider";
import { Block, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { AptosEvent } from "../../../domain/entities/aptos";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "winston";

export class AptosJsonRPCBlockRepository {
  private readonly logger: winston.Logger;

  constructor(private readonly client: InstrumentedAptosProvider) {
    this.logger = winston.child({ module: "AptosJsonRPCBlockRepository" });
  }

  async getSequenceNumber(
    range: Block | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    try {
      const fromBlock = range?.fromBlock ? Number(range?.fromBlock) : undefined;
      const toSequence = range?.toBlock ? Number(range?.toBlock) : undefined;

      const results = await this.client.getEventsByEventHandle(
        filter.address,
        filter.event!,
        filter.fieldName,
        fromBlock,
        toSequence
      );
      return results;
    } catch (e) {
      this.handleError(e, "getSequenceNumber");
      throw e;
    }
  }

  async getTransactionsByVersionForSourceEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<TransactionsByVersion[]> {
    try {
      const transactions = await Promise.all(
        events.map(async (event) => {
          const transaction = await this.client.getTransactionByVersion(Number(event.version));
          const block = await this.client.getBlockByVersion(Number(event.version));

          const wormholeEvent = transaction.events.find((tx: any) => tx.type === filter.type);

          return {
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
        })
      );

      return transactions;
    } catch (e) {
      this.handleError(e, "getTransactionsByVersionForSourceEvent");
      throw e;
    }
  }

  async getTransactionsByVersionForRedeemedEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<TransactionsByVersion[]> {
    try {
      const transactions = await Promise.all(
        events.map(async (event) => {
          const transaction = await this.client.getTransactionByVersion(Number(event.version));
          const block = await this.client.getBlockByVersion(Number(event.version));

          const vaaBuffer = Buffer.from(transaction.payload.arguments[0].substring(2), "hex");
          const vaa = parseVaa(vaaBuffer);

          return {
            consistencyLevel: vaa.consistencyLevel,
            emitterChain: vaa.emitterChain,
            blockHeight: block.block_height,
            timestamp: vaa.timestamp,
            blockTime: vaa.timestamp,
            sequence: vaa.sequence,
            version: transaction.version,
            payload: vaa.payload.toString("hex"),
            address: filter.address,
            sender: vaa.emitterAddress.toString("hex"),
            status: transaction.success,
            events: transaction.events,
            nonce: vaa.nonce,
            hash: transaction.hash,
            type: filter.type,
          };
        })
      );

      return transactions;
    } catch (e) {
      this.handleError(e, "getTransactionsByVersionForRedeemedEvent");
      throw e;
    }
  }

  async getTransactions(block: Block): Promise<any[]> {
    try {
      const results = await this.client.getTransactions(block);
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
  emitterChain?: number;
  blockHeight: bigint;
  timestamp: number;
  blockTime: number;
  sequence: bigint;
  version: string;
  payload: string;
  address: string;
  sender: string;
  status?: boolean;
  events: any;
  nonce: number;
  hash: string;
  type?: string;
};
