import { SNSClient } from "@aws-sdk/client-sns";
import { Config } from "./config";
import { SnsEventRepository } from "./repositories/SnsEventRepository";
import axios, { AxiosInstance } from "axios";
import axiosRateLimit from "axios-rate-limit";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./repositories/EvmJsonRPCBlockRepository";
import { FileMetadataRepo } from "./repositories/FileMetadataRepo";

export class RepositoriesBuilder {
  private cfg: Config;
  private snsClient?: SNSClient;
  private axiosInstance?: AxiosInstance;
  private repositories = new Map();

  constructor(cfg: Config) {
    this.cfg = cfg;
    this.build();
  }

  private build() {
    this.snsClient = new SNSClient({ region: this.cfg.sns.region });
    this.axiosInstance = axiosRateLimit(axios.create(), {
      perMilliseconds: 1000,
      maxRequests: 1_000,
    }); // TODO: configurable per repo

    this.repositories.set("sns", new SnsEventRepository(this.snsClient, this.cfg.sns));
    this.cfg.metadata?.dir &&
      this.repositories.set("metadata", new FileMetadataRepo(this.cfg.metadata.dir));
    this.cfg.supportedChains.forEach((chain) => {
      const repoCfg: EvmJsonRPCBlockRepositoryCfg = {
        chain,
        rpc: this.cfg.platforms[chain].rpcs[0],
        timeout: this.cfg.platforms[chain].timeout,
      };
      this.repositories.set(
        `${chain}-evmRepo`,
        new EvmJsonRPCBlockRepository(repoCfg, this.axiosInstance!)
      );
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

  close(): void {
    this.snsClient?.destroy();
  }
}
