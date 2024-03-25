import { TransactionFoundEvent } from "../../entities";
import { StatRepository } from "../../repositories";
import { WormchainLog } from "../../entities/wormchain";

export class HandleWormchainLogs {
  constructor(
    private readonly cfg: HandleWormchainLogsOptions,
    private readonly mapper: (tx: WormchainLog) => TransactionFoundEvent[],
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(logs: WormchainLog[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    logs.map((log) => {
      const logMapped = this.mapper(log);

      if (logMapped && logMapped.length > 0) {
        logMapped.forEach((log) => {
          this.report();
          filterLogs.push(log);
        });
      }
      return logMapped;
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
  metricLabels?: { job: string; chain: string; commitment: string };
  eventTypes?: string[];
  metricName: string;
  id: string;
}
