import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import {
  CosmosTransaction,
  IbcTransaction,
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

  getBlockHeight(chainId: number): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(chainId)).execute();
  }

  getBlockTransactions(
    chainId: number,
    blockNumbers: Set<bigint>,
    attributesTypes: string[]
  ): Promise<CosmosTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getBlockTransactions(chainId, blockNumbers, attributesTypes))
      .execute();
  }

  getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]> {
    return this.breaker.fn(() => this.delegate.getRedeems(ibcTransaction)).execute();
  }
}
