import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { CosmosTransaction } from "../../../domain/entities/cosmos";
import { CosmosRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import { Filter } from "../../../domain/actions/cosmos/types";
import winston from "winston";

export class RateLimitedCosmosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<CosmosRepository>
  implements CosmosRepository
{
  constructor(delegate: CosmosRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedCosmosJsonRPCBlockRepository" });
  }

  getBlockTimestamp(blockNumber: bigint, chain: string): Promise<number | undefined> {
    return this.breaker.fn(() => this.delegate.getBlockTimestamp(blockNumber, chain)).execute();
  }

  getTransactions(
    filter: Filter,
    blockBatchSize: number,
    chain: string
  ): Promise<CosmosTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactions(filter, blockBatchSize, chain))
      .execute();
  }
}
