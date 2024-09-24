import { Checkpoint, SuiEventFilter, TransactionFilter } from "@mysten/sui.js/client";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { SuiRepository } from "../../../domain/repositories";
import { Range } from "../../../domain/entities";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedSuiJsonRPCBlockRepository
  extends RateLimitedRPCRepository<SuiRepository>
  implements SuiRepository
{
  constructor(
    delegate: SuiRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedSuiJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<void> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
  }

  getLastCheckpointNumber(): Promise<bigint> {
    return this.breaker.fn(() => this.delegate.getLastCheckpointNumber()).execute();
  }

  getCheckpoint(sequence: string | number | bigint): Promise<Checkpoint> {
    return this.breaker.fn(() => this.delegate.getCheckpoint(sequence)).execute();
  }

  getLastCheckpoint(): Promise<Checkpoint> {
    return this.breaker.fn(() => this.delegate.getLastCheckpoint()).execute();
  }

  getCheckpoints(range: Range): Promise<Checkpoint[]> {
    return this.breaker.fn(() => this.delegate.getCheckpoints(range)).execute();
  }

  getTransactionBlockReceipts(digests: string[]): Promise<SuiTransactionBlockReceipt[]> {
    return this.breaker.fn(() => this.delegate.getTransactionBlockReceipts(digests)).execute();
  }

  queryTransactions(
    filter?: TransactionFilter | undefined,
    cursor?: string | undefined
  ): Promise<SuiTransactionBlockReceipt[]> {
    return this.breaker.fn(() => this.delegate.queryTransactions(filter, cursor)).execute();
  }

  queryTransactionsByEvent(
    filter: SuiEventFilter,
    cursor?: string | undefined
  ): Promise<SuiTransactionBlockReceipt[]> {
    return this.breaker.fn(() => this.delegate.queryTransactionsByEvent(filter, cursor)).execute();
  }
}
