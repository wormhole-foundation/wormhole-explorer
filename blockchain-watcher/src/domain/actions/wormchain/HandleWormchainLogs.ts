import { TransactionFoundEvent } from "../../entities";
import { StatRepository } from "../../repositories";
import { WormchainLog } from "../../entities/wormchain";

export class HandleWormchainLogs {
  constructor(
    private readonly cfg: HandleWormchainLogsOptions,
    private readonly mapper: (tx: WormchainLog) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(logs: WormchainLog[]): Promise<TransactionFoundEvent[]> {
    const mappedItems = logs.map((log) => {
      const logMap = this.mapper(log);
      return logMap;
    });

    const filterItems = mappedItems.filter((item) => {
      if (item) {
        this.report();
        return item;
      }
    });

    await this.target(filterItems);
    return filterItems;
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
  metricLabels?: { job: string; chain: string; commitment: string };
  eventTypes?: string[];
  metricName: string;
  id: string;
}
