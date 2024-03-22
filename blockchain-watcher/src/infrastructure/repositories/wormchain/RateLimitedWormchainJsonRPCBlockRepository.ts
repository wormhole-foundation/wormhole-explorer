import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedWormchainJsonRPCBlockRepository
  extends RateLimitedRPCRepository<WormchainRepository>
  implements WormchainRepository
{
  constructor(delegate: WormchainRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedWormchainJsonRPCBlockRepository" });
  }

  getBlockHeight(finality: string): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(finality)).execute();
  }

  getBlockLogs(blockNumber: bigint): Promise<any> {
    return this.breaker.fn(() => this.delegate.getBlockLogs(blockNumber)).execute();
  }
}
