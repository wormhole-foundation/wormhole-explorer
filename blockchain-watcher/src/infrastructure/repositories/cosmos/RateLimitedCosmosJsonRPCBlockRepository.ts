import { CosmosRepository, ProviderHealthCheck } from "../../../domain/repositories";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { CosmosTransaction } from "../../../domain/entities/cosmos";
import { Options } from "../common/rateLimitedOptions";
import { Filter } from "../../../domain/actions/cosmos/types";
import winston from "winston";

export class RateLimitedCosmosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<CosmosRepository>
  implements CosmosRepository
{
  constructor(
    delegate: CosmosRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedCosmosJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
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
