import { TransactionFoundEvent } from "../../entities";
import { CosmosRedeem } from "../../entities/wormchain";
import { StatRepository } from "../../repositories";

export class HandleCosmosRedeems {
  constructor(
    private readonly cfg: HandleCosmosRedeemsOptions,
    private readonly mapper: (
      addresses: string[],
      cosmosRedeem: CosmosRedeem
    ) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(cosmosRedeem: CosmosRedeem[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    cosmosRedeem.forEach((redeem) => {
      const redeemMapped = this.mapper(this.cfg.filter.addresses, redeem);

      if (redeemMapped) {
        this.report(redeemMapped.attributes.protocol, redeemMapped.attributes.chain!);
        filterLogs.push(redeemMapped);
      }
    });

    await this.target(filterLogs);
    return filterLogs;
  }

  private report(protocol: string, chain: string) {
    const labels = {
      commitment: "immediate",
      job: this.cfg.id,
      protocol,
      chain,
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleCosmosRedeemsOptions {
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
