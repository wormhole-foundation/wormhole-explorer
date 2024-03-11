import { AptosEvent, AptosTransaction } from "../../../domain/entities/aptos";
import { Block, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { AptosRepository } from "../../../domain/repositories";
import { Options } from "../common/rateLimitedOptions";
import winston from "winston";

export class RateLimitedAptosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<AptosRepository>
  implements AptosRepository
{
  constructor(delegate: AptosRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedAptosJsonRPCBlockRepository" });
  }

  getSequenceNumber(range: Block | undefined, filter: TransactionFilter): Promise<AptosEvent[]> {
    return this.breaker.fn(() => this.delegate.getSequenceNumber(range, filter)).execute();
  }

  getTransactionsByVersionForSourceEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<AptosTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactionsByVersionForSourceEvent(events, filter))
      .execute();
  }

  getTransactionsByVersionForRedeemedEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<AptosTransaction[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactionsByVersionForRedeemedEvent(events, filter))
      .execute();
  }

  getTransactions(range: Block | undefined): Promise<any[]> {
    return this.breaker.fn(() => this.delegate.getTransactions(range)).execute();
  }
}
