import { Commitment } from "@solana/web3.js";
import { Circuit, Ratelimit, RatelimitError } from "mollitia";
import { solana } from "../../../domain/entities";
import { SolanaSlotRepository } from "../../../domain/repositories";
import { Fallible, SolanaFailure, ErrorType } from "../../../domain/errors";

export class RateLimitedSolanaSlotRepository implements SolanaSlotRepository {
  delegate: SolanaSlotRepository;
  breaker: Circuit;

  constructor(delegate: SolanaSlotRepository, opts: Options = { period: 10_000, limit: 50 }) {
    this.delegate = delegate;
    this.breaker = new Circuit({
      options: {
        modules: [new Ratelimit({ limitPeriod: opts.period, limitForPeriod: opts.limit })],
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
      return result;
    } catch (err) {
      if (err instanceof RatelimitError) {
        return Fallible.error(new SolanaFailure(0, err.message, ErrorType.Ratelimit));
      }

      return Fallible.error(new SolanaFailure(err, "unknown error"));
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
