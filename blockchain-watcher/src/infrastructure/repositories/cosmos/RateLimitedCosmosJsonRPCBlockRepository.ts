import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { CosmosRepository } from "../../../domain/repositories";
import { CosmosRedeem } from "../../../domain/entities/wormchain";
import { Options } from "../common/rateLimitedOptions";
import { Filter } from "../../../domain/actions/cosmos/types";
import winston from "winston";

export class RateLimitedCosmosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<CosmosRepository>
  implements CosmosRepository
{
  constructor(delegate: CosmosRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedSeiJsonRPCBlockRepository" });
  }

  getBlockTimestamp(
    blockNumber: bigint,
    chainId: number,
    chain: string
  ): Promise<number | undefined> {
    return this.breaker
      .fn(() => this.delegate.getBlockTimestamp(blockNumber, chainId, chain))
      .execute();
  }

  getRedeems(
    chainId: number,
    filter: Filter,
    blockBatchSize: number,
    chain: string
  ): Promise<CosmosRedeem[]> {
    return this.breaker
      .fn(() => this.delegate.getRedeems(chainId, filter, blockBatchSize, chain))
      .execute();
  }
}
