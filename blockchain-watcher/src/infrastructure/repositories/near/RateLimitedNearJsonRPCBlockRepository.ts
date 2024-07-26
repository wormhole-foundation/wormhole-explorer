import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { NearRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedNearJsonRPCBlockRepository
  extends RateLimitedRPCRepository<NearRepository>
  implements NearRepository
{
  constructor(delegate: NearRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedNearJsonRPCBlockRepository" });
  }

  getBlockHeight(): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight()).execute();
  }
}
