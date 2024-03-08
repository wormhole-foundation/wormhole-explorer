import { Block, TransactionFilter } from "../../../domain/actions/aptos/PollAptos";
import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { TransactionsByVersion } from "./AptosJsonRPCBlockRepository";
import { AptosRepository } from "../../../domain/repositories";
import { AptosEvent } from "../../../domain/entities/aptos";
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

  getTransactionsByVersionsForSourceEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<TransactionsByVersion[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactionsByVersionsForSourceEvent(events, filter))
      .execute();
  }

  getTransactionsByVersionsForRedeemedEvent(
    events: AptosEvent[],
    filter: TransactionFilter
  ): Promise<TransactionsByVersion[]> {
    return this.breaker
      .fn(() => this.delegate.getTransactionsByVersionsForRedeemedEvent(events, filter))
      .execute();
  }

  getTransactions(range: Block | undefined): Promise<any[]> {
    return this.breaker.fn(() => this.delegate.getTransactions(range)).execute();
  }
}
