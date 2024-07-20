import { TransactionFoundEvent } from "../../entities";
import { CosmosTransaction } from "../../entities/cosmos";
import { StatRepository } from "../../repositories";

export class HandleCosmosTransactions {
  constructor(
    private readonly cfg: HandleCosmosTransactionsOptions,
    private readonly mapper: (
      addresses: string[],
      cosmosTransactions: CosmosTransaction
    ) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(cosmosTransactions: CosmosTransaction[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    cosmosTransactions.forEach((tx) => {
      const transactionMapped = this.mapper(this.cfg.filter.addresses, tx);

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

export interface HandleCosmosTransactionsOptions {
  metricLabels?: { job: string; chain: string; commitment: string };
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
