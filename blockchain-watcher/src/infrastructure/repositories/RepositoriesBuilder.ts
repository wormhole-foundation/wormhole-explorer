import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import pg from "pg";
import { Connection } from "@solana/web3.js";
import { Config, DBConfig } from "../config";
import {
  SnsEventRepository,
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  FileMetadataRepository,
  PromStatRepository,
  StaticJobRepository,
  Web3SolanaSlotRepository,
  RateLimitedSolanaSlotRepository,
  BscEvmJsonRPCBlockRepository,
  PostgresMetadataRepository,
  PostgresJobExecutionRepository,
  InMemoryJobExecutionRepository,
  ArbitrumEvmJsonRPCBlockRepository,
} from "./";
import { HttpClient } from "../rpc/http/HttpClient";
import {
  JobExecutionRepository,
  JobRepository,
  MetadataRepository,
} from "../../domain/repositories";
import { MoonbeamEvmJsonRPCBlockRepository } from "./evm/MoonbeamEvmJsonRPCBlockRepository";

const SOLANA_CHAIN = "solana";
const EVM_CHAIN = "evm";
const EVM_CHAINS = new Map([
  ["ethereum", "evmRepo"],
  ["avalanche", "evmRepo"],
  ["oasis", "evmRepo"],
  ["fantom", "evmRepo"],
  ["karura", "evmRepo"],
  ["acala", "evmRepo"],
  ["klaytn", "evmRepo"],
  ["celo", "evmRepo"],
  ["optimism", "evmRepo"],
  ["base", "evmRepo"],
  ["bsc", "bsc-evmRepo"],
  ["arbitrum", "arbitrum-evmRepo"],
  ["moonbeam", "moonbeam-evmRepo"],
]);

export class RepositoriesBuilder {
  private cfg: Config;
  private snsClient?: SNSClient;
  private closeables: (() => Promise<void>)[] = [];
  private repositories = new Map();

  constructor(cfg: Config) {
    this.cfg = cfg;
  }

  public async init(): Promise<void> {
    this.snsClient = this.createSnsClient();

    await this.loadMetadataRepositories();

    this.repositories.set("sns", new SnsEventRepository(this.snsClient, this.cfg.sns));

    this.repositories.set("metrics", new PromStatRepository());

    this.cfg.enabledPlatforms.forEach((chain) => {
      if (chain === SOLANA_CHAIN) {
        const cfg = this.cfg.chains[chain];
        const solanaSlotRepository = new RateLimitedSolanaSlotRepository(
          new Web3SolanaSlotRepository(
            new Connection(cfg.rpcs[0], { disableRetryOnRateLimit: true })
          ),
          cfg.rateLimit
        );
        this.repositories.set("solana-slotRepo", solanaSlotRepository);
      }

      if (chain === EVM_CHAIN) {
        const httpClient = this.createHttpClient();
        const repoCfg: EvmJsonRPCBlockRepositoryCfg = {
          chains: this.cfg.chains,
        };
        this.repositories.set("bsc-evmRepo", new BscEvmJsonRPCBlockRepository(repoCfg, httpClient));
        this.repositories.set("evmRepo", new EvmJsonRPCBlockRepository(repoCfg, httpClient));
        this.repositories.set(
          "moonbeam-evmRepo",
          new MoonbeamEvmJsonRPCBlockRepository(repoCfg, httpClient)
        );
        this.repositories.set(
          "arbitrum-evmRepo",
          new ArbitrumEvmJsonRPCBlockRepository(repoCfg, httpClient, this.getMetadataRepository())
        );
      }
    });

    this.repositories.set(
      "job-executions",
      this.cfg.jobExecutions.use === "postgres"
        ? await this.createPostgresJobExecutionRepository(this.cfg.dbConfig)
        : new InMemoryJobExecutionRepository()
    );
    this.repositories.set(
      "jobs",
      new StaticJobRepository(
        this.cfg.jobs.dir,
        this.cfg.dryRun,
        (chain: string) => this.getEvmBlockRepository(chain),
        {
          metadataRepo: this.getMetadataRepository(),
          statsRepo: this.getStatsRepository(),
          snsRepo: this.getSnsEventRepository(),
          solanaSlotRepo: this.getSolanaSlotRepository(),
        }
      )
    );
  }

  public getEvmBlockRepository(chain: string): EvmJsonRPCBlockRepository {
    const instanceRepoName = EVM_CHAINS.get(chain);
    if (!instanceRepoName) throw new Error(`Chain ${chain} not supported`);
    return this.getRepo(instanceRepoName);
  }

  public getSnsEventRepository(): SnsEventRepository {
    return this.getRepo("sns");
  }

  public getMetadataRepository(): MetadataRepository<any> {
    return this.getRepo("metadata");
  }

  public getStatsRepository(): PromStatRepository {
    return this.getRepo("metrics");
  }

  public getJobsRepository(): JobRepository {
    return this.getRepo("jobs");
  }

  public getJobExecutionRepository(): JobExecutionRepository {
    return this.getRepo("job-executions");
  }

  public getSolanaSlotRepository(): Web3SolanaSlotRepository {
    return this.getRepo("solana-slotRepo");
  }

  private getRepo(name: string): any {
    const repo = this.repositories.get(name);
    if (!repo) throw new Error(`No repository ${name}`);

    return repo;
  }

  public async close() {
    this.snsClient?.destroy();
    this.closeables.forEach(async (closeable) => await closeable());
  }

  private async loadMetadataRepositories() {
    if (
      this.cfg.metadata.use.includes("fs") &&
      this.cfg.metadata.use.includes("postgres") &&
      this.cfg.metadata.dir &&
      this.cfg.dbConfig
    ) {
      this.repositories.set(
        "metadata",
        new FileMetadataRepository(
          this.cfg.metadata.dir,
          await this.createPostgresMetadataRepository(this.cfg.dbConfig)
        )
      );
      return;
    }

    if (this.cfg.metadata.use.includes("fs") && this.cfg.metadata.dir) {
      this.repositories.set("metadata", new FileMetadataRepository(this.cfg.metadata.dir));
      return;
    }

    if (this.cfg.metadata.use.includes("postgres") && this.cfg.dbConfig) {
      this.repositories.set(
        "metadata",
        await this.createPostgresMetadataRepository(this.cfg.dbConfig)
      );
    }

    this.getMetadataRepository();
  }

  private createSnsClient(): SNSClient {
    const snsCfg: SNSClientConfig = { region: this.cfg.sns.region };
    if (this.cfg.sns.credentials) {
      snsCfg.credentials = {
        accessKeyId: this.cfg.sns.credentials.accessKeyId,
        secretAccessKey: this.cfg.sns.credentials.secretAccessKey,
      };
      snsCfg.endpoint = this.cfg.sns.credentials.url;
    }

    const client = new SNSClient(snsCfg);
    this.closeables.push(async () => client.destroy());
    return client;
  }

  private async createPgPool(dbCfg: DBConfig): Promise<pg.Pool> {
    const pgPool = new pg.Pool({
      connectionString: dbCfg.connString,
      connectionTimeoutMillis: dbCfg.connectionTimeout ?? 30_000,
      query_timeout: dbCfg.queryTimeout ?? 20_000,
      max: dbCfg.maxPoolSize ?? 10,
      options: `-c search_path=${this.cfg.environment}`,
    });
    return pgPool;
  }

  private async createPostgresMetadataRepository(
    dbCfg: DBConfig
  ): Promise<PostgresMetadataRepository> {
    const repo = new PostgresMetadataRepository(
      await this.createPgPool(dbCfg),
      this.cfg.environment
    );
    await repo.init();
    this.closeables.push(async () => repo.close());
    return repo;
  }

  private async createPostgresJobExecutionRepository(
    dbCfg?: DBConfig
  ): Promise<PostgresJobExecutionRepository> {
    if (!dbCfg) {
      throw new Error("Missing db config");
    }
    const repo = new PostgresJobExecutionRepository(
      await this.createPgPool({ ...dbCfg, maxPoolSize: 1 }) // maxPoolSize should always be 1 here
    );
    this.closeables.push(async () => repo.close());
    return repo;
  }

  private createHttpClient(): HttpClient {
    return new HttpClient({
      retries: 3,
      timeout: 1_0000,
      initialDelay: 1_000,
      maxDelay: 30_000,
    });
  }
}
