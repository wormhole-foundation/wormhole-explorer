import { Checkpoint, SuiEventFilter, TransactionFilter } from "@mysten/sui.js/client";
import { Circuit, Ratelimit, Retry, RetryMode } from "mollitia";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { SuiRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import { Range } from "../../../domain/entities";
import winston from "winston";

export class RateLimitedSuiJsonRPCBlockRepository implements SuiRepository {
  private delegate: SuiRepository;
  private breaker: Circuit;
  private logger: winston.Logger = winston.child({
    module: "RateLimitedSuiJsonRPCBlockRepository",
  });

  constructor(delegate: SuiRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    this.delegate = delegate;
    this.breaker = new Circuit({
      options: {
        modules: [
          new Ratelimit({ limitPeriod: opts.period, limitForPeriod: opts.limit }),
          new Retry({
            attempts: 2,
            interval: 1_000,
            fastFirst: false,
            mode: RetryMode.EXPONENTIAL,
            factor: 1,
            onRejection: (err: Error | any) => {
              if (err.message?.startsWith("429 Too Many Requests")) {
                this.logger.warn("Got 429 from sui RPC node. Retrying in 10 secs...");
                return 10_000; // Wait 10 secs if we get a 429
              } else {
                return true; // Retry according to config
              }
            },
          }),
        ],
      },
    });
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
