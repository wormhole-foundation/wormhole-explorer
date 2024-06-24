import { TransactionFoundEvent } from "../../entities";
import { AlgorandTransaction } from "../../entities/algorand";
import { StatRepository } from "../../repositories";

export class HandleAlgorandTransactions {
  constructor(
    private readonly cfg: HandleAlgorandTransactionsOptions,
    private readonly mapper: (tx: AlgorandTransaction, filter: Filter[]) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(txs: AlgorandTransaction[]): Promise<TransactionFoundEvent[]> {
    const items: TransactionFoundEvent[] = [];

    for (const tx of txs) {
      const txMapped = this.mapper(tx, this.cfg.filter);
      if (txMapped) {
        this.report(txMapped.attributes.protocol);
        items.push(txMapped);
      }
    }

    await this.target(items);

    return items;
  }

  private report(protocol: string) {
    const labels = {
      job: this.cfg.id,
      chain: "algorand",
      protocol: protocol ?? "unknown",
      commitment: "latest",
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleAlgorandTransactionsOptions {
  metricName: string;
  filter: Filter[];
  id: string;
}

type Filter = {
  applicationsIds: string;
  applicationAddress: string;
};
