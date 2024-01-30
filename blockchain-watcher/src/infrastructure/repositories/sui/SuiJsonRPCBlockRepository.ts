import {
  Checkpoint,
  SuiClient,
  SuiEventFilter,
  SuiTransactionBlockResponse,
  TransactionFilter,
} from "@mysten/sui.js/client";
import winston from "winston";
import { Range } from "../../../domain/entities";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { SuiRepository } from "../../../domain/repositories";
import { divideIntoBatches } from "../common/utils";

const QUERY_MAX_RESULT_LIMIT_CHECKPOINTS = 100;
const TX_BATCH_SIZE = 50;

export class SuiJsonRPCBlockRepository implements SuiRepository {
  private readonly client: SuiClient;
  private readonly logger: winston.Logger;

  constructor(private readonly cfg: SuiJsonRPCBlockRepositoryConfig) {
    this.client = new SuiClient({ url: this.cfg.rpc });
    this.logger = winston.child({ module: "SuiJsonRPCBlockRepository" });
    this.logger.info(`[sui][SuiJsonRPCBlockRepository] Using RPC node ${this.cfg.rpc}`);
  }

  async getLastCheckpointNumber(): Promise<bigint> {
    try {
      const res = await this.client.getLatestCheckpointSequenceNumber();
      return BigInt(res);
    } catch (e) {
      this.handleError(e, "getLatestCheckpointNumber");
      throw e;
    }
  }

  async getCheckpoints(range: Range): Promise<Checkpoint[]> {
    const count = Number(range.to - range.from + 1n);
    const checkpoints = [...new Array(count).keys()].map((i) =>
      (range.from + BigInt(i)).toString()
    );

    const batches = divideIntoBatches(new Set(checkpoints), QUERY_MAX_RESULT_LIMIT_CHECKPOINTS);

    const results: Checkpoint[] = [];
    for (const batch of batches) {
      let res;
      try {
        res = await this.client.getCheckpoints({
          cursor: (BigInt(Array.from(batch)[0]) - 1n).toString(),
          descendingOrder: false,
          limit: Math.min(count, QUERY_MAX_RESULT_LIMIT_CHECKPOINTS),
        });
      } catch (e) {
        this.handleError(e, "getCheckpoints");
        throw e;
      }

      for (const checkpoint of res.data) {
        results.push(checkpoint);
      }
    }

    return results;
  }

  async getTransactionBlockReceipts(digests: string[]): Promise<SuiTransactionBlockReceipt[]> {
    const batches = divideIntoBatches(new Set(digests), TX_BATCH_SIZE);

    let receipts: SuiTransactionBlockResponse[] = [];
    for (const batch of batches) {
      let res;
      try {
        res = await this.client.multiGetTransactionBlocks({
          digests: Array.from(batch),
          options: { showEvents: true, showInput: true },
        });
      } catch (e) {
        this.handleError(e, "multiGetTransactionBlocks");
        throw e;
      }

      for (const tx of res) {
        receipts.push(tx);
      }
    }

    return receipts.map(this.mapTransactionBlockReceipt);
  }

  private mapTransactionBlockReceipt(tx: SuiTransactionBlockResponse): SuiTransactionBlockReceipt {
    return {
      ...tx,
      digest: tx.digest,
      checkpoint: tx.checkpoint!,
      timestampMs: tx.timestampMs!,
      transaction: tx.transaction!,
      events: tx.events || [],
      errors: tx.errors || [],
    };
  }

  async queryTransactions(
    filter?: TransactionFilter,
    cursor?: string | undefined
  ): Promise<SuiTransactionBlockReceipt[]> {
    let response;
    try {
      response = await this.client.queryTransactionBlocks({
        filter,
        order: "ascending",
        cursor,
        options: {
          showEvents: true,
          showInput: true,
        },
      });
    } catch (e) {
      this.handleError(e, "queryTransactions");
      throw e;
    }

    return response.data.map(this.mapTransactionBlockReceipt);
  }

  async queryTransactionsByEvent(
    query: SuiEventFilter,
    cursor?: string | undefined
  ): Promise<SuiTransactionBlockReceipt[]> {
    let response;
    try {
      response = await this.client.queryEvents({
        query,
        order: "ascending",
        cursor: cursor ? { txDigest: cursor, eventSeq: "0" } : undefined,
        limit: TX_BATCH_SIZE,
      });
    } catch (e) {
      this.handleError(e, "queryTransactionsByEvent");
      throw e;
    }

    const txs = response.data.map((e) => e.id.txDigest);

    return this.getTransactionBlockReceipts(txs);
  }

  async getCheckpoint(id: string | bigint | number): Promise<Checkpoint> {
    try {
      return this.client.getCheckpoint({ id: id.toString() });
    } catch (e) {
      this.handleError(e, "getCheckpoint");
      throw e;
    }
  }

  async getLastCheckpoint(): Promise<Checkpoint> {
    const id = await this.getLastCheckpointNumber();
    return this.getCheckpoint(id.toString());
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[sui][SuiJsonRPCBlockRepository] Error calling ${method}: ${e.message} (rpc ${this.cfg.rpc})`);
  }
}

export type SuiJsonRPCBlockRepositoryConfig = {
  rpc: string;
};
