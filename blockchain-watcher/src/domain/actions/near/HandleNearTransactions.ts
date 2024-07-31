import { TransactionFoundEvent } from "../../entities";
import { NearTransaction } from "../../entities/near";
import { StatRepository } from "../../repositories";

export class HandleNearTransactions {
  constructor(
    private readonly cfg: HandleNearTransactionsOptions,
    private readonly mapper: (nearTransactions: NearTransaction) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(nearTransactions: NearTransaction[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    nearTransactions.forEach((tx) => {
      const transactionMapped = this.mapper(tx);

      if (transactionMapped) {
        this.report(transactionMapped.attributes.protocol, transactionMapped.attributes.chain!);
        filterLogs.push(transactionMapped);
      }
    });

    await this.target(filterLogs);
    return filterLogs;
  }

  private report(protocol: string, chain: string) {
    if (!this.cfg.metricName) return;

    const labels = this.cfg.metricLabels ?? {
      commitment: "immediate",
      job: this.cfg.id,
      protocol,
      chain,
    };

    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleNearTransactionsOptions {
  metricLabels?: { job: string; chain: string; commitment: string };
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
