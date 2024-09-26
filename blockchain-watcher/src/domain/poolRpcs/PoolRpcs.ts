import { StatRepository } from "../repositories";
import { RunPoolRpcs } from "../actions/RunPoolRpcs";
import winston from "winston";

export class PoolRpcs extends RunPoolRpcs {
  protected readonly logger: winston.Logger;

  private cfg: PoolRpcsConfig;
  private reportValues: any[] = [];

  private readonly statsRepo: StatRepository;
  private readonly metadataRepo: any;

  constructor(statsRepo: StatRepository, metadataRepo: any, cfg: PoolRpcsConfig) {
    super(statsRepo);
    this.logger = winston.child({ module: "PoolRpcs", label: "pool-rpcs" });
    this.statsRepo = statsRepo;
    this.metadataRepo = metadataRepo;
    this.cfg = cfg;
  }

  protected async set(): Promise<void> {
    try {
      const promises = this.cfg.getProps().map(async (cfg) => {
        const { id, repository, chain, commitment } = cfg;

        let normalizeCursor;
        const metadata = await this.metadataRepo.get(id);

        if (metadata) {
          normalizeCursor = this.normalizeCursor(metadata);
        }

        if (!repository) {
          this.logger.error(`Repository not found: [chain: ${chain} - repository: ${repository}]`);
          return;
        }

        const result = await repository.healthCheck(chain, commitment, normalizeCursor);
        this.reportValues.push({
          rpcs: result,
          commitment,
          chain,
          id,
        });
      });

      await Promise.all(promises);
    } catch (e) {
      this.logger.error(`Error setting providers: ${e}`);
    }
  }

  protected report(): void {
    for (const report of this.reportValues) {
      let labels: Label = {
        job: `pool-rpcs-${report.id}`,
        chain: report.chain,
        commitment: report.commitment,
      };

      for (const rpc of report.rpcs) {
        labels.rpc = rpc.url;
        this.statsRepo.measure("pool_rpc_latency", rpc.latency, {
          ...labels,
        });

        this.statsRepo.measure("pool_rpc_height", rpc.height, {
          ...labels,
        });
      }
    }
  }

  private normalizeCursor(blockHeight: { [key: string]: any }): string {
    const keys = ["lastBlock", "lastFrom", "lastSlot", "lastCursor"];
    let height;

    for (const key of keys) {
      if (blockHeight.hasOwnProperty(key)) {
        height = blockHeight[key];
        if (height && height["checkpoint"]) {
          height = height["checkpoint"];
        }
        break;
      }
    }
    return height;
  }
}

type Label = {
  commitment: string;
  chain: string;
  job: string;
  rpc?: string;
};

export type ProviderHealthCheck = {
  isHealthy: boolean;
  latency?: number;
  height: bigint | undefined;
  url: string;
};

export interface PoolRpcsConfigProps {
  environment: string;
  commitment: string;
  repository: any;
  interval?: number;
  chainId: number;
  chain: string;
  id: string;
}

export class PoolRpcsConfig {
  private props: PoolRpcsConfigProps[];

  constructor(props: PoolRpcsConfigProps[]) {
    this.props = props;
  }

  public getProps(): PoolRpcsConfigProps[] {
    return this.props;
  }
}
