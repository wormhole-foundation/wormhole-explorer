import { Circuit, Ratelimit, Retry, RetryMode } from "mollitia";
import { EvmBlockRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "../../log";
import {
  EvmBlock,
  EvmLogFilter,
  EvmLog,
  EvmTag,
  ReceiptTransaction,
} from "../../../domain/entities";

export class RateLimitedEvmJsonRPCBlockRepository implements EvmBlockRepository {
  private delegate: EvmBlockRepository;
  private breaker: Circuit;
  private logger: winston.Logger = winston.child({
    module: "RateLimitedEvmJsonRPCBlockRepository",
  });

  constructor(delegate: EvmBlockRepository, opts: Options = { period: 10_000, limit: 50 }) {
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
                this.logger.warn("Got 429 from evm RPC node. Retrying in 10 secs...");
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

  getBlockHeight(chain: string, finality: string): Promise<bigint> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(chain, finality)).execute();
  }

  getBlocks(chain: string, blockNumbers: Set<bigint>): Promise<Record<string, EvmBlock>> {
    return this.breaker.fn(() => this.delegate.getBlocks(chain, blockNumbers)).execute();
  }

  getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]> {
    return this.breaker.fn(() => this.delegate.getFilteredLogs(chain, filter)).execute();
  }

  getTransactionReceipt(
    chain: string,
    hashNumbers: Set<string>
  ): Promise<Record<string, ReceiptTransaction>> {
    return this.breaker.fn(() => this.delegate.getTransactionReceipt(chain, hashNumbers)).execute();
  }

  getBlock(
    chain: string,
    blockNumberOrTag: bigint | EvmTag,
    isTransactionsPresent: boolean
  ): Promise<EvmBlock> {
    return this.breaker
      .fn(() => this.delegate.getBlock(chain, blockNumberOrTag, isTransactionsPresent))
      .execute();
  }
}
