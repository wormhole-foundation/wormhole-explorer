import { TransactionFoundEvent } from "../../entities";
import { CosmosTransaction } from "../../entities/wormchain";
import { StatRepository } from "../../repositories";

export class HandleWormchainTransactions {
  constructor(
    private readonly cfg: HandleWormchainLogsOptions,
    private readonly mapper: (
      addresses: string[],
      tx: CosmosTransaction
    ) => TransactionFoundEvent[],
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(transactions: CosmosTransaction[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    transactions.forEach((transaction) => {
      const logMapped = this.mapper(this.cfg.filter.addresses, transaction);

      if (logMapped.length > 0) {
        logMapped.forEach((log) => {
          this.report();
          filterLogs.push(log);
        });
      }
    });

    await this.target(filterLogs);
    return filterLogs;
  }

  private report() {
    const labels = {
      commitment: "immediate",
      chain: "wormchain",
      job: this.cfg.id,
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleWormchainLogsOptions {
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
