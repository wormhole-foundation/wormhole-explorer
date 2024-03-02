import { TransactionFoundEvent } from "../../entities";
import { StatRepository } from "../../repositories";
import { AptosEvent } from "../../entities/aptos";

export class HandleAptosTransactions {
  constructor(
    private readonly cfg: HandleAptosTransactionsOptions,
    private readonly mapper: (tx: AptosEvent) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(txs: AptosEvent[]): Promise<TransactionFoundEvent[]> {
    const items: TransactionFoundEvent[] = [];
    await this.target(items);

    return items;
  }

  private report(protocol: string) {
    if (!this.cfg.metricName) return;

    const labels = this.cfg.metricLabels ?? {
      job: this.cfg.id,
      chain: "aptos",
      commitment: "finalized",
      protocol: protocol ?? "unknown",
    };

    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleAptosTransactionsOptions {
  metricLabels?: { job: string; chain: string; commitment: string };
  eventTypes?: string[];
  metricName?: string;
  id: string;
}
