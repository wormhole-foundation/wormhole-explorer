import { TransactionFoundEvent } from "../../entities";
import { WormchainBlockLogs } from "../../entities/wormchain";
import { StatRepository } from "../../repositories";

export class HandleWormchainLogs {
  constructor(
    private readonly cfg: HandleWormchainLogsOptions,
    private readonly mapper: (
      addresses: string[],
      tx: WormchainBlockLogs
    ) => TransactionFoundEvent[],
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(logs: WormchainBlockLogs[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    logs.forEach((log) => {
      const logMapped = this.mapper(this.cfg.filter.addresses, log);

      if (logMapped.length > 0) {
        logMapped.forEach((log) => {
          if (log) {
            this.report();
            filterLogs.push(log);
          }
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
