import { RunPoolRpcs } from "../RunPoolRpcs";
import { Repos } from "../../../infrastructure/repositories";
import winston from "winston";

export class PoolRpcs extends RunPoolRpcs {
  protected readonly logger: winston.Logger;

  private repositories: Repos;
  private cfg: PoolRpcsConfig;
  private reportValues: any;

  constructor(repositories: Repos, cfg: PoolRpcsConfig) {
    super(repositories);
    this.logger = winston.child({ module: "PoolRpcs", label: "pool-rpcs" });
    this.repositories = repositories;
    this.cfg = cfg;
  }

  protected async set(): Promise<void> {
    try {
      for (const cfg of this.cfg.getProps()) {
        const { id, repository, chain, commitment } = cfg;

        const metadata = await this.repositories.metadataRepo.get(id);
        const cursor = this.normalizeCursor(metadata);

        if (!repository) {
          this.logger.error(`Repository not found: ${repository}`);
          continue;
        }

        const result = await repository.healthCheck(chain, commitment, cursor);
        this.reportValues = {
          id,
          chain,
          commitment,
          ...result,
        };
      }
    } catch (e) {
      this.logger.error(`Error setting providers: ${e}`);
    }
  }

  protected report(): void {
    const labels = {
      job: `pool-rpcs-${this.reportValues.id}`,
      chain: this.reportValues.chain,
      commitment: this.reportValues.commitment,
    };
  }

  private normalizeCursor(blockHeight: { [key: string]: any }): string {
    const keys = ["lastBlock", "lastFrom", "lastSlot", "lastCursor"];
    let height;

    for (const key of keys) {
      if (blockHeight.hasOwnProperty(key)) {
        height = blockHeight[key];
        break;
      }
    }
    return height;
  }
}

export type ProviderHealthCheck = {
  latency?: number;
  height: bigint | undefined;
  isLive: boolean;
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
