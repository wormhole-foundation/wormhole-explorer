import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import { IbcTransaction, CosmosRedeem } from "../../../domain/entities/wormchain";

export class RateLimitedWormchainJsonRPCBlockRepository
  extends RateLimitedRPCRepository<WormchainRepository>
  implements WormchainRepository
{
  constructor(delegate: WormchainRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedWormchainJsonRPCBlockRepository" });
  }

  getTxs(chainId: number, address: string, blockBatchSize: number): Promise<any[]> {
    return this.breaker.fn(() => this.delegate.getTxs(chainId, address, blockBatchSize)).execute();
  }

  getBlockTimestamp(chainId: number, blockNumber: bigint): Promise<number | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockTimestamp(chainId, blockNumber)).execute();
  }

  getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]> {
    return this.breaker.fn(() => this.delegate.getRedeems(ibcTransaction)).execute();
  }
}
