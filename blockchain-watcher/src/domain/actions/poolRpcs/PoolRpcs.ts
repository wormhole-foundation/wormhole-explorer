import { RunPoolRpcs } from "../RunPoolRpcs";
import { Repos } from "../../../infrastructure/repositories";
import winston from "winston";

export class PoolRpcs extends RunPoolRpcs {
  protected readonly logger: winston.Logger;

  private repositories: Repos;
  private cfg: PoolRpcsConfig;

  constructor(repositories: Repos, cfg: PoolRpcsConfig) {
    super(repositories);
    this.logger = winston.child({ module: "PoolRpcs", label: "pool-rpcs" });
    this.repositories = repositories;
    this.cfg = cfg;
  }

  protected async set(): Promise<void> {
    setInterval(async () => {
      try {
        const metadata = await this.repositories.metadataRepo.get(this.cfg.id);
        const blockHeightCursor = this.normalizeBlockHeightCursor(metadata);
        const repository = this.cfg.repository;
        const chain = this.cfg.chain;

        const repo =
          repository == "evmRepo"
            ? this.repositories.evmRepo(chain)
            : (this.repositories[repository as keyof Repos] as any);
        if (!repo) {
          this.logger.error(`Repository not found: ${repository}`);
          return;
        }

        const pool = await repo.getPool(chain);
        const providers = pool.getProviders();
        const heights = await repo.getAllBlockHeight(providers, this.cfg.commitment);

        if (heights || heights.length > 0) {
          await pool.setProviders(providers, heights, blockHeightCursor);
        }
      } catch (e) {
        this.logger.error(`Error setting providers: ${e}`);
      }
    }, 5 * 60 * 1000); // 5 minutes
  }

  protected report(): void {}

  private normalizeBlockHeightCursor(blockHeight: { [key: string]: any }): string {
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
  repository: string;
  interval?: number;
  chainId: number;
  chain: string;
  id: string;
}

export type ProviderHeight = { url: string; height: bigint };

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

  public get repository() {
    return this.props.repository;
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
