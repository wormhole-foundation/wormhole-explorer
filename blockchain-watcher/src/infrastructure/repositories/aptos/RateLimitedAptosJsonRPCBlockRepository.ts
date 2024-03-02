import { RateLimitedRPCRepository } from "../RateLimitedRPCRepository";
import { TransactionsByVersion } from "./AptosJsonRPCBlockRepository";
import { AptosRepository } from "../../../domain/repositories";
import { AptosEvent } from "../../../domain/entities/aptos";
import { Options } from "../common/rateLimitedOptions";
import { Sequence } from "../../../domain/actions/aptos/PollAptosTransactions";
import winston from "winston";

export class RateLimitedAptosJsonRPCBlockRepository
  extends RateLimitedRPCRepository<AptosRepository>
  implements AptosRepository
{
  constructor(delegate: AptosRepository, opts: Options = { period: 10_000, limit: 1000 }) {
    super(delegate, opts);
    this.logger = winston.child({ module: "RateLimitedAptosJsonRPCBlockRepository" });
  }

  getSequenceNumber(range: Sequence | undefined): Promise<AptosEvent[]> {
    return this.breaker.fn(() => this.delegate.getSequenceNumber(range)).execute();
  }

  getTransactionsForVersions(events: AptosEvent[]): Promise<TransactionsByVersion[]> {
    return this.breaker.fn(() => this.delegate.getTransactionsForVersions(events)).execute();
  }
}
