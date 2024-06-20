import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { AlgorandRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedAlgorandJsonRPCBlockRepository
  extends RateLimitedRPCRepository<AlgorandRepository>
  implements AlgorandRepository
{
  constructor(delegate: AlgorandRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedAlgorandJsonRPCBlockRepository" });
  }

  getApplicationsLogs(address: string, fromBlock: bigint, toBlock: bigint): Promise<any[]> {
    return this.breaker
      .fn(() => this.delegate.getApplicationsLogs(address, fromBlock, toBlock))
      .execute();
  }

  getBlockHeight(): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight()).execute();
  }
}
