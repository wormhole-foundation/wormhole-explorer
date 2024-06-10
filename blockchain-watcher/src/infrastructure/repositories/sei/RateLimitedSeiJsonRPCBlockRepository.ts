import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { SeiRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import { SeiRedeem } from "../../../domain/entities/sei";
import winston from "winston";

export class RateLimitedSeiJsonRPCBlockRepository
  extends RateLimitedRPCRepository<SeiRepository>
  implements SeiRepository
{
  constructor(delegate: SeiRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedSeiJsonRPCBlockRepository" });
  }

  getBlockTimestamp(blockNumber: bigint): Promise<number | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockTimestamp(blockNumber)).execute();
  }

  getRedeems(chainId: number, address: string): Promise<SeiRedeem[]> {
    return this.breaker.fn(() => this.delegate.getRedeems(chainId, address)).execute();
  }
}
