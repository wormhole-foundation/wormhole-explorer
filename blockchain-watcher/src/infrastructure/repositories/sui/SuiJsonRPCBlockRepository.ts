import { Checkpoint, SuiClient, SuiTransactionBlockResponse } from "@mysten/sui.js/client";
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
    this.logger.info(`Using RPC node ${this.cfg.rpc}`);
  }

  async getLastCheckpoint(): Promise<bigint> {
    const res = await this.client.getLatestCheckpointSequenceNumber();
    return BigInt(res);
  }

  // TODO: handle case where range length is larger than max limit
  async getCheckpoints(range: Range): Promise<Checkpoint[]> {
    const count = Number(range.to - range.from + 1n);

    const res = await this.client.getCheckpoints({
      cursor: (range.from - 1n).toString(),
      descendingOrder: false,
      limit: Math.min(count, QUERY_MAX_RESULT_LIMIT_CHECKPOINTS),
    });

    return res.data;
  }

  async getTransactionBlockReceipts(digests: string[]): Promise<SuiTransactionBlockReceipt[]> {
    const batches = divideIntoBatches(new Set(digests), TX_BATCH_SIZE);

    let receipts: SuiTransactionBlockResponse[] = [];
    for (const batch of batches) {
      const res = await this.client.multiGetTransactionBlocks({
        digests: Array.from(batch),
        options: { showEvents: true, showInput: true },
      });

      for (const tx of res) {
        receipts.push(tx);
      }
    }

    return receipts.map((tx) => ({
      ...tx,
      digest: tx.digest,
      checkpoint: tx.checkpoint!,
      timestampMs: tx.timestampMs!,
      transaction: tx.transaction!,
      events: tx.events || [],
      errors: tx.errors || [],
    }));
  }
}

export type SuiJsonRPCBlockRepositoryConfig = {
  rpc: string;
};
