import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import { EvmBlockRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import {
  ReceiptTransaction,
  EvmLogFilter,
  EvmBlock,
  EvmLog,
  EvmTag,
} from "../../../domain/entities";

export class RateLimitedEvmJsonRPCBlockRepository
  extends RateLimitedRPCRepository<EvmBlockRepository>
  implements EvmBlockRepository
{
  constructor(
    delegate: EvmBlockRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedEvmJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: EvmTag, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
  }

  getBlockHeight(chain: string, finality: string): Promise<bigint> {
    return this.breaker.fn(() => this.delegate.getBlockHeight(chain, finality)).execute();
  }

  getBlocks(
    chain: string,
    blockNumbers: Set<bigint>,
    isTransactionsPresent: boolean
  ): Promise<Record<string, EvmBlock>> {
    return this.breaker
      .fn(() => this.delegate.getBlocks(chain, blockNumbers, isTransactionsPresent))
      .execute();
  }

  getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]> {
    return this.breaker.fn(() => this.delegate.getFilteredLogs(chain, filter)).execute();
  }

  getTransactionReceipt(
    chain: string,
    hashNumbers: Set<string>
  ): Promise<Record<string, ReceiptTransaction>> {
    return this.breaker.fn(() => this.delegate.getTransactionReceipt(chain, hashNumbers)).execute();
  }

  getBlock(
    chain: string,
    blockNumberOrTag: bigint | EvmTag,
    isTransactionsPresent: boolean
  ): Promise<EvmBlock> {
    return this.breaker
      .fn(() => this.delegate.getBlock(chain, blockNumberOrTag, isTransactionsPresent))
      .execute();
  }
}
