import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { Config } from "./config";
import {
  SnsEventRepository,
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  FileMetadataRepo,
  PromStatRepository,
} from "./repositories";

import { HttpClient } from "./repositories/HttpClient";

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
      const httpClient = this.createHttpClient(this.cfg.platforms[chain].timeout);
      const repoCfg: EvmJsonRPCBlockRepositoryCfg = {
        chain,
        rpc: this.cfg.platforms[chain].rpcs[0],
        timeout: this.cfg.platforms[chain].timeout,
      };
      this.repositories.set(`${chain}-evmRepo`, new EvmJsonRPCBlockRepository(repoCfg, httpClient));
    });
  }

  public getEvmBlockRepository(chain: string): EvmJsonRPCBlockRepository {
    const repo = this.repositories.get(`${chain}-evmRepo`);
    if (!repo) throw new Error(`No EvmJsonRPCBlockRepository for chain ${chain}`);

    return repo;
  }

  public getSnsEventRepository(): SnsEventRepository {
    const repo = this.repositories.get("sns");
    if (!repo) throw new Error(`No SnsEventRepository`);

    return repo;
  }

  public getMetadataRepository(): FileMetadataRepo {
    const repo = this.repositories.get("metadata");
    if (!repo) throw new Error(`No FileMetadataRepo`);

    return repo;
  }

  public getStatsRepository(): PromStatRepository {
    const repo = this.repositories.get("metrics");
    if (!repo) throw new Error(`No PromStatRepository`);

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
