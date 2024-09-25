import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import { AlgorandRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedAlgorandJsonRPCBlockRepository
  extends RateLimitedRPCRepository<AlgorandRepository>
  implements AlgorandRepository
{
  constructor(
    delegate: AlgorandRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedAlgorandJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
  }

  getTransactions(
    applicationId: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<AlgorandTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactions(applicationId, fromBlock, toBlock))
      .execute();
  }

  getBlockHeight(): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight()).execute();
  }
}
