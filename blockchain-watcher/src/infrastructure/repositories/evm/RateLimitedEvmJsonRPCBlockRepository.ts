import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { EvmBlockRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";
import {
  EvmBlock,
  EvmLogFilter,
  EvmLog,
  EvmTag,
  ReceiptTransaction,
} from "../../../domain/entities";

export class RateLimitedEvmJsonRPCBlockRepository
  extends RateLimitedRPCRepository<EvmBlockRepository>
  implements EvmBlockRepository
{
  constructor(delegate: EvmBlockRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedEvmJsonRPCBlockRepository" });
  }

  getTransactionByHash(chain: string, hashNumbers: Set<string>): Promise<any> {
    return this.breaker.fn(() => this.delegate.getTransactionByHash(chain, hashNumbers)).execute();
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
