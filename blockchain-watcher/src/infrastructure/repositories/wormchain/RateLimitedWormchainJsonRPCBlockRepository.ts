import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import {
  CosmosTransactionByWormchain,
  WormchainBlockLogs,
  CosmosRedeem,
} from "../../../domain/entities/wormchain";

export class RateLimitedWormchainJsonRPCBlockRepository
  extends RateLimitedRPCRepository<WormchainRepository>
  implements WormchainRepository
{
  constructor(delegate: WormchainRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedWormchainJsonRPCBlockRepository" });
  }

  getBlockHeight(): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight()).execute();
  }

  getBlockLogs(
    chainId: number,
    blockNumber: bigint,
    filterTypes: string[]
  ): Promise<WormchainBlockLogs> {
    return this.breaker
      .fn(() => this.delegate.getBlockLogs(chainId, blockNumber, filterTypes))
      .execute();
  }

  getRedeems(cosmosTransactionByWormchain: CosmosTransactionByWormchain): Promise<CosmosRedeem[]> {
    return this.breaker.fn(() => this.delegate.getRedeems(cosmosTransactionByWormchain)).execute();
  }
}
