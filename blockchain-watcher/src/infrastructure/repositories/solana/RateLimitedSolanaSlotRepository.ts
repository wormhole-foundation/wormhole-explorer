import { SolanaSlotRepository, ProviderHealthCheck } from "../../../domain/repositories";
import { Fallible, SolanaFailure, ErrorType } from "../../../domain/errors";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { RatelimitError } from "mollitia";
import { Options } from "../common/rateLimitedOptions";
import { solana } from "../../../domain/entities";
import winston from "../../../infrastructure/log";

export class RateLimitedSolanaSlotRepository
  extends RateLimitedRPCRepository<SolanaSlotRepository>
  implements SolanaSlotRepository
{
  constructor(
    delegate: SolanaSlotRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 50, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedSolanaSlotRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
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
