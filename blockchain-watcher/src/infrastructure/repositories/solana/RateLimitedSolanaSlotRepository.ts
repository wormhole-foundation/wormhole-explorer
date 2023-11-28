import { Circuit, Ratelimit, RatelimitError, Retry, RetryMode } from "mollitia";
import { solana } from "../../../domain/entities";
import { SolanaSlotRepository } from "../../../domain/repositories";
import { Fallible, SolanaFailure, ErrorType } from "../../../domain/errors";
import winston from "../../../infrastructure/log";

export class RateLimitedSolanaSlotRepository implements SolanaSlotRepository {
  delegate: SolanaSlotRepository;
  breaker: Circuit;
  logger: winston.Logger = winston.child({ module: "RateLimitedSolanaSlotRepository" });

  constructor(delegate: SolanaSlotRepository, opts: Options = { period: 10_000, limit: 50 }) {
    this.delegate = delegate;
    this.breaker = new Circuit({
      options: {
        modules: [
          new Ratelimit({ limitPeriod: opts.period, limitForPeriod: opts.limit }),
          new Retry({
            attempts: 1,
            interval: 10_000,
            fastFirst: false,
            mode: RetryMode.LINEAR,
            factor: 1,
            onRejection: (err: Error | any) => {
              if (err.message?.startsWith("429 Too Many Requests")) {
                this.logger.warn("Got 429 from solana RPC node. Retrying in 10 secs...");
                return 10_000; // Wait 10 secs if we get a 429
              } else {
                return false; // Dont retry, let the caller handle it
              }
            },
          }),
        ],
      },
    });
  }

  getLatestSlot(commitment: string): Promise<number> {
    return this.breaker.fn(() => this.delegate.getLatestSlot(commitment)).execute();
  }

  async getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, SolanaFailure>> {
    try {
      const result: Fallible<solana.Block, SolanaFailure> = await this.breaker
        .fn(() => this.delegate.getBlock(slot, finality))
        .execute();

      if (!result.isOk()) {
        throw result.getError();
      }

      return result;
    } catch (err: SolanaFailure | any) {
      // this needs more handling due to delegate.getBlock returning a Fallible with a SolanaFailure
      if (err instanceof RatelimitError) {
        return Fallible.error(new SolanaFailure(0, err.message, ErrorType.Ratelimit));
      }

      if (err instanceof SolanaFailure) {
        return Fallible.error(err);
      }

      return Fallible.error(new SolanaFailure(err, err?.message ?? "unknown error"));
    }
  }

  getSignaturesForAddress(
    address: string,
    beforeSig: string,
    afterSig: string,
    limit: number
  ): Promise<solana.ConfirmedSignatureInfo[]> {
    return this.breaker
      .fn(() => this.delegate.getSignaturesForAddress(address, beforeSig, afterSig, limit))
      .execute(address, beforeSig, afterSig, limit);
  }

  getTransactions(sigs: solana.ConfirmedSignatureInfo[]): Promise<solana.Transaction[]> {
    return this.breaker.fn(() => this.delegate.getTransactions(sigs)).execute(sigs);
  }
}

export type Options = {
  period: number;
  limit: number;
};
