import { TransactionFoundEvent } from "../../entities";
import { StatRepository } from "../../repositories";
import { CosmosRedeem } from "../../entities/wormchain";
import { mapChain } from "../../../common/wormchain";

export class HandleWormchainRedeems {
  constructor(
    private readonly cfg: HandleWormchainRedeemsOptions,
    private readonly mapper: (cosmosRedeem: CosmosRedeem) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>,
    private readonly statsRepo: StatRepository
  ) {}

  public async handle(cosmosRedeems: CosmosRedeem[]): Promise<TransactionFoundEvent[]> {
    const filterLogs: TransactionFoundEvent[] = [];

    cosmosRedeems.forEach((redeem) => {
      const redeemMapped = this.mapper(redeem);

      if (redeemMapped) {
        this.report(redeemMapped.attributes.protocol, redeemMapped.chainId);
        filterLogs.push(redeemMapped);
      }
    });

    await this.target(filterLogs);
    return filterLogs;
  }

  private report(protocol: string, chainId: number) {
    const labels = {
      commitment: "immediate",
      chain: mapChain(chainId),
      job: this.cfg.id,
      protocol,
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }
}

export interface HandleWormchainRedeemsOptions {
  metricName: string;
  filter: { addresses: string[] };
  id: string;
}
