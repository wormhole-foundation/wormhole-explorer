import { TransactionFoundEvent } from "../../entities";
import { TransactionsByVersion } from "../../../infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { StatRepository } from "../../repositories";

export class HandleAptosTransactions {
  constructor(
    private readonly cfg: HandleAptosTransactionsOptions,
    private readonly mapper: (tx: TransactionsByVersion) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(txs: TransactionsByVersion[]): Promise<TransactionFoundEvent[]> {
    const items: TransactionFoundEvent[] = [];

    for (const tx of txs) {
      const txMapped = this.mapper(tx);
      this.report(txMapped.attributes.protocol);
      items.push(txMapped);
    }

    await this.target(items);

    return items;
  }

  private report(protocol: string) {
    if (!this.cfg.metricName) return;

    const labels = this.cfg.metricLabels ?? {
      job: this.cfg.id,
      chain: "aptos",
      commitment: "finalized",
      protocol: protocol,
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
