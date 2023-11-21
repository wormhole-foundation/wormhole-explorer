import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { Connection } from "@solana/web3.js";
import { Config } from "./config";
import {
  SnsEventRepository,
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  FileMetadataRepo,
  PromStatRepository,
  StaticJobRepository,
  Web3SolanaSlotRepository,
} from "./repositories";

import { HttpClient } from "./repositories/HttpClient";
import { JobRepository } from "../domain/repositories";

export class RepositoriesBuilder {
  private cfg: Config;
  private snsClient?: SNSClient;
  private repositories = new Map();

  constructor(cfg: Config) {
    this.cfg = cfg;
    this.build();
  }

  private build() {
    this.snsClient = this.createSnsClient();

    this.repositories.set("sns", new SnsEventRepository(this.snsClient, this.cfg.sns));
    this.repositories.set("metrics", new PromStatRepository());

    this.cfg.metadata?.dir &&
      this.repositories.set("metadata", new FileMetadataRepo(this.cfg.metadata.dir));

    this.cfg.supportedChains.forEach((chain) => {
      if (!this.cfg.platforms[chain]) throw new Error(`No config for chain ${chain}`);

      if (chain === "solana") {
        const cfg = this.cfg.platforms[chain];
        const solanaSlotRepository = new Web3SolanaSlotRepository(new Connection(cfg.rpcs[0]));
        this.repositories.set("solana-slotRepo", solanaSlotRepository);
      }

      if (chain === "ethereum") {
        const httpClient = this.createHttpClient(this.cfg.platforms[chain].timeout);
        const repoCfg: EvmJsonRPCBlockRepositoryCfg = {
          chain,
          rpc: this.cfg.platforms[chain].rpcs[0],
          timeout: this.cfg.platforms[chain].timeout,
        };
        this.repositories.set(
          `${chain}-evmRepo`,
          new EvmJsonRPCBlockRepository(repoCfg, httpClient)
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
        }
      )
    );
  }

  public getEvmBlockRepository(chain: string): EvmJsonRPCBlockRepository {
    return this.getRepo(`${chain}-evmRepo`);
  }

  public getSnsEventRepository(): SnsEventRepository {
    return this.getRepo("sns");
  }

  public getMetadataRepository(): FileMetadataRepo {
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

  private createHttpClient(timeout?: number, retries?: number): HttpClient {
    return new HttpClient({
      retries: retries ?? 3,
      timeout: timeout ?? 5_000,
      initialDelay: 1_000,
      maxDelay: 30_000,
    });
  }
}
