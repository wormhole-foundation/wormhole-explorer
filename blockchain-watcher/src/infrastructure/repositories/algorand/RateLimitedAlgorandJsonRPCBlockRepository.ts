import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
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
    opts: Options = { period: 10_000, limit: 1000 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedAlgorandJsonRPCBlockRepository" });
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
