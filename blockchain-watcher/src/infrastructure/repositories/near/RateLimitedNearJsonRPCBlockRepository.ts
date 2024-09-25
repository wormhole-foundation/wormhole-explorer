import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import { NearTransaction } from "../../../domain/entities/near";
import { NearRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedNearJsonRPCBlockRepository
  extends RateLimitedRPCRepository<NearRepository>
  implements NearRepository
{
  constructor(
    delegate: NearRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedNearJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
  }

  getBlockHeight(commitment: string): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(commitment)).execute();
  }

  getTransactions(
    contract: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<NearTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactions(contract, fromBlock, toBlock))
      .execute();
  }
}
