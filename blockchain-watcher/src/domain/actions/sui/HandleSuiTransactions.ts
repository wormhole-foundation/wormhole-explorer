import { TransactionFoundEvent } from "../../entities";
import { SuiTransactionBlockReceipt } from "../../entities/sui";
import { StatRepository } from "../../repositories";

export class HandleSuiTransactions {
  constructor(
    private readonly cfg: HandleSuiTransactionsOptions,
    private readonly mapper: (tx: SuiTransactionBlockReceipt) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(txs: SuiTransactionBlockReceipt[]): Promise<TransactionFoundEvent[]> {
    const items: TransactionFoundEvent[] = [];

    for (const tx of txs) {
      const valid = this.filterTransaction(tx);
      if (valid) {
        this.report();
        items.push(this.mapper(tx));
      }
    }

    await this.target(items);

    return items;
  }

  private filterTransaction({ events }: SuiTransactionBlockReceipt): boolean {
    return (
      !!events && !!this.cfg.eventTypes && events.some((e) => this.cfg.eventTypes!.includes(e.type))
    );
  }

  private report() {
    if (!this.cfg.metricName) return;

    const labels = {
      job: this.cfg.id,
      chain: "sui",
      commitment: "immediate",
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleSuiTransactionsOptions {
  eventTypes?: string[];
  metricName?: string;
  id: string;
}
