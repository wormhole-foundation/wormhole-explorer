import { TransactionFoundEvent } from "../../entities";
import { StatRepository } from "../../repositories";
import { SeiRedeem } from "../../entities/sei";

export class HandleSeiRedeems {
  constructor(
    private readonly cfg: HandleSeiRedeemsOptions,
    private readonly mapper: (addresses: string[], seiRedeem: SeiRedeem) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(seiRedeem: SeiRedeem[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    seiRedeem.forEach((redeem) => {
      const redeemMapped = this.mapper(this.cfg.filter.addresses, redeem);

      if (redeemMapped) {
        this.report(redeemMapped.attributes.protocol);
        filterLogs.push(redeemMapped);
      }
    });

    await this.target(filterLogs);
    return filterLogs;
  }

  private report(protocol: string) {
    const labels = {
      commitment: "immediate",
      chain: "sei",
      job: this.cfg.id,
      protocol,
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleSeiRedeemsOptions {
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
