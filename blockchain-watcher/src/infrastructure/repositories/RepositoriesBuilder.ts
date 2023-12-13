import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { Connection } from "@solana/web3.js";
import { Config } from "../config";
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
  ArbitrumEvmJsonRPCBlockRepository,
  PolygonJsonRPCBlockRepository,
  MoonbeamEvmJsonRPCBlockRepository,
} from ".";
import { HttpClient } from "../rpc/http/HttpClient";
import { JobRepository } from "../../domain/repositories";

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
  ["polygon", "polygon-evmRepo"],
]);

export class RepositoriesBuilder {
  private cfg: Config;
  private snsClient?: SNSClient;
  private repositories = new Map();

  constructor(cfg: Config) {
    this.cfg = cfg;
    this.build();
  }

  private build(): void {
    this.snsClient = this.createSnsClient();
    this.repositories.set("sns", new SnsEventRepository(this.snsClient, this.cfg.sns));

    this.repositories.set("metrics", new PromStatRepository());

    this.cfg.metadata?.dir &&
      this.repositories.set("metadata", new FileMetadataRepository(this.cfg.metadata.dir));

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
          "polygon-evmRepo",
          new PolygonJsonRPCBlockRepository(repoCfg, httpClient)
        );
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

  public getMetadataRepository(): FileMetadataRepository {
    return this.getRepo("metadata");
  }

  public getStatsRepository(): PromStatRepository {
    return this.getRepo("metrics");
  }

  public getJobsRepository(): JobRepository {
    return this.getRepo("jobs");
  }

  public getSolanaSlotRepository(): Web3SolanaSlotRepository {
    return this.getRepo("solana-slotRepo");
  }

  private getRepo(name: string): any {
    const repo = this.repositories.get(name);
    if (!repo) throw new Error(`No repository ${name}`);

    return repo;
  }

  public close(): void {
    this.snsClient?.destroy();
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

    return new SNSClient(snsCfg);
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
