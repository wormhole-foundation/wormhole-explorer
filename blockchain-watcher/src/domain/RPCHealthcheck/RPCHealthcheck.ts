import { MetadataRepository, StatRepository } from "../repositories";
import { RunRPCHealthcheck } from "../actions/RunRPCHealthcheck";
import winston from "winston";

export class RPCHealthcheck extends RunRPCHealthcheck {
  protected readonly logger: winston.Logger;

  private cfg: RPCHealthcheckConfig[];
  private reportValues: any[] = [];

  private readonly statsRepo: StatRepository;
  private readonly metadataRepo: any;

  constructor(
    statsRepo: StatRepository,
    metadataRepo: MetadataRepository<any>,
    cfg: RPCHealthcheckConfig[],
    interval: number
  ) {
    super(statsRepo, interval);
    this.logger = winston.child({ module: "RunRPCHealthcheck", label: "rpc-healthcheck" });
    this.statsRepo = statsRepo;
    this.metadataRepo = metadataRepo;
    this.cfg = cfg;
  }

  protected async execute(): Promise<void> {
    try {
      const promises = this.cfg.map(async (cfg) => {
        const { id, repository, chain, commitment } = cfg;
        if (!repository) {
          this.logger.error(`Repository not found: [chain: ${chain} - repository: ${repository}]`);
          return;
        }

        let normalizeCursor;
        const metadata = await this.metadataRepo.get(id);

        if (metadata) {
          normalizeCursor = this.normalizeCursor(metadata);
        }

        const result = await repository.healthCheck(chain, commitment, normalizeCursor);
        return {
          rpcsStatus: result,
          chain,
          id,
        };
      });

      const results = await Promise.allSettled(promises);

      results.forEach((result) => {
        if (result.status === "fulfilled" && result?.value) {
          this.reportValues.push(result.value);
        } else if (result.status === "rejected") {
          this.logger.error(`Promise rejected: ${result.reason}`);
        }
      });
    } catch (e) {
      this.logger.error(`Error setting providers: ${e}`);
    }
  }

  protected report(): void {
    for (const report of this.reportValues) {
      let labels: Label = {
        commitment: report.commitment,
        chain: report.chain,
        job: `rpc-healthcheck-${report.id}`,
      };

      for (const rpc of report.rpcsStatus) {
        labels.rpc = rpc.url;
        this.statsRepo.measure("rpc_healthcheck_latency", rpc.latency, {
          ...labels,
        });

        this.statsRepo.measure("rpc_healthcheck_height", rpc.height, {
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

export interface RPCHealthcheckConfig {
  environment: string;
  commitment: string;
  repository: any;
  interval?: number;
  chainId: number;
  chain: string;
  id: string;
}
[];
