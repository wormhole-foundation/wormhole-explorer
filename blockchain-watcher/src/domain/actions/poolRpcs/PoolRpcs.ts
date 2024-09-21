import { RunPoolRpcs } from "../RunPoolRpcs";
import winston from "winston";

export class PoolRpcs extends RunPoolRpcs {
  protected readonly logger: winston.Logger;

  private repositories: Map<string | string[], any>;
  private cfg: PoolRpcsConfig;

  constructor(repositories: Map<string | string[], any>, cfg: PoolRpcsConfig) {
    super(repositories);
    this.repositories = repositories;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PoolRpcs", label: "pool-rpcs" });
  }

  protected async set(): Promise<void> {
    setInterval(async () => {
      try {
        const metadata = await this.repositories.get("metadata").get(this.cfg.id);
        const height = this.normalizeBlockData(metadata);
        const chain = this.cfg.chain;

        const repo = this.repositories.get("evmRepo");

        if (repo) {
          await repo.setProviders(chain, this.cfg.commitment, height);
          const a = repo(chain);
        }
        await repo.setProviders();
      } catch (e) {}
    }, 5 * 1000); // 5 minutes
  }

  protected report(): void {}

  private normalizeBlockData(blockHeight: { [key: string]: any }): string {
    const keys = ["lastBlock", "blockHeight", "latestBlock", "currentBlock"];
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

export interface PoolRpcsConfigProps {
  environment: string;
  commitment: string;
  interval?: number;
  chainId: number;
  chain: string;
  id: string;
}

export class PoolRpcsConfig {
  private props: PoolRpcsConfigProps;

  constructor(props: PoolRpcsConfigProps) {
    this.props = props;
  }

  public get commitment() {
    return this.props.commitment;
  }

  public get interval() {
    return this.props.interval;
  }

  public get id() {
    return this.props.id;
  }

  public get chain() {
    return this.props.chain;
  }

  public get environment() {
    return this.props.environment;
  }

  public get chainId() {
    return this.props.chainId;
  }
}
