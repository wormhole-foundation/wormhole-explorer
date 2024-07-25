import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import {
  IbcTransaction,
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

  getBlockHeight(chain: string): Promise<bigint | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(chain)).execute();
  }

  getBlockLogs(
    chain: string,
    blockNumber: bigint,
    attributesTypes: string[]
  ): Promise<WormchainBlockLogs> {
    return this.breaker
      .fn(() => this.delegate.getBlockLogs(chain, blockNumber, attributesTypes))
      .execute();
  }

  getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]> {
    return this.breaker.fn(() => this.delegate.getRedeems(ibcTransaction)).execute();
  }
}
