import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { WormchainRepository } from "../../../domain/repositories";
import { ProviderHealthCheck } from "../../../domain/actions/poolRpcs/PoolRpcs";
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
  constructor(
    delegate: WormchainRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedWormchainJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
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
