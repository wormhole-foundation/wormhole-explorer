import { Fallible, SolanaFailure, ErrorType } from "../../../domain/errors";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { SolanaSlotRepository } from "../../../domain/repositories";
import { RatelimitError } from "mollitia";
import { Options } from "../common/rateLimitedOptions";
import { solana } from "../../../domain/entities";
import winston from "../../../infrastructure/log";

export class RateLimitedSolanaSlotRepository
  extends RateLimitedRPCRepository<SolanaSlotRepository>
  implements SolanaSlotRepository
{
  constructor(delegate: SolanaSlotRepository, opts: Options = { period: 10_000, limit: 50 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedSolanaSlotRepository" });
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
    limit: number,
    finality?: string
  ): Promise<solana.ConfirmedSignatureInfo[]> {
    return this.breaker
      .fn(() =>
        this.delegate.getSignaturesForAddress(address, beforeSig, afterSig, limit, finality)
      )
      .execute();
  }

  getTransactions(
    sigs: solana.ConfirmedSignatureInfo[],
    finality?: string
  ): Promise<solana.Transaction[]> {
    return this.breaker.fn(() => this.delegate.getTransactions(sigs, finality)).execute();
  }
}
