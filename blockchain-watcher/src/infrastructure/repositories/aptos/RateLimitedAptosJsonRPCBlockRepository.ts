import { AptosEvent, AptosTransaction } from "../../../domain/entities/aptos";
import { Range, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { ProviderHealthCheck } from "../../../domain/actions/poolRpcs/PoolRpcs";
import { AptosRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedAptosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<AptosRepository>
  implements AptosRepository
{
  constructor(
    delegate: AptosRepository,
    chain: string,
    opts: Options = { period: 10_000, limit: 1000, interval: 1_000, attempts: 10 }
  ) {
    super(delegate, chain, opts);
    this.logger = winston.child({ module: "RateLimitedAptosJsonRPCBlockRepository" });
  }

  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]> {
    return this.breaker.fn(() => this.delegate.healthCheck(chain, finality, cursor)).execute();
  }

  getTransactions(range: Range | undefined): Promise<any[]> {
    return this.breaker.fn(() => this.delegate.getTransactions(range)).execute();
  }

  getEventsByEventHandle(
    range: Range | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]> {
    return this.breaker.fn(() => this.delegate.getEventsByEventHandle(range, filter)).execute();
  }

  getTransactionsByVersion(records: AptosEvent[]): Promise<AptosTransaction[]> {
    return this.breaker.fn(() => this.delegate.getTransactionsByVersion(records)).execute();
  }
}
