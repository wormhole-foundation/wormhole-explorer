import { RateLimitedWormchainJsonRPCBlockRepository } from "./wormchain/RateLimitedWormchainJsonRPCBlockRepository";
import { RateLimitedAptosJsonRPCBlockRepository } from "./aptos/RateLimitedAptosJsonRPCBlockRepository";
import { RateLimitedEvmJsonRPCBlockRepository } from "./evm/RateLimitedEvmJsonRPCBlockRepository";
import { RateLimitedSuiJsonRPCBlockRepository } from "./sui/RateLimitedSuiJsonRPCBlockRepository";
import { WormchainJsonRPCBlockRepository } from "./wormchain/WormchainJsonRPCBlockRepository";
import { AptosJsonRPCBlockRepository } from "./aptos/AptosJsonRPCBlockRepository";
import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { InstrumentedHttpProvider } from "../rpc/http/InstrumentedHttpProvider";
import { Config } from "../config";
import {
  WormchainRepository,
  AptosRepository,
  JobRepository,
  SuiRepository,
} from "../../domain/repositories";
import {
  InstrumentedConnection,
  InstrumentedSuiClient,
  providerPoolSupplier,
  RpcConfig,
} from "@xlabs/rpc-pool";
import {
  MoonbeamEvmJsonRPCBlockRepository,
  ArbitrumEvmJsonRPCBlockRepository,
  RateLimitedSolanaSlotRepository,
  PolygonJsonRPCBlockRepository,
  BscEvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  EvmJsonRPCBlockRepository,
  SuiJsonRPCBlockRepository,
  Web3SolanaSlotRepository,
  FileMetadataRepository,
  StaticJobRepository,
  PromStatRepository,
  SnsEventRepository,
  ProviderPoolMap,
} from ".";

const WORMCHAIN_CHAIN = "wormchain";
const SOLANA_CHAIN = "solana";
const APTOS_CHAIN = "aptos";
const EVM_CHAIN = "evm";
const SUI_CHAIN = "sui";
const EVM_CHAINS = new Map([
  ["ethereum", "evmRepo"],
  ["ethereum-sepolia", "evmRepo"],
  ["avalanche", "evmRepo"],
  ["oasis", "evmRepo"],
  ["fantom", "evmRepo"],
  ["karura", "evmRepo"],
  ["acala", "evmRepo"],
  ["klaytn", "evmRepo"],
  ["celo", "evmRepo"],
  ["optimism", "evmRepo"],
  ["optimism-sepolia", "evmRepo"],
  ["base", "evmRepo"],
  ["base-sepolia", "evmRepo"],
  ["bsc", "bsc-evmRepo"],
  ["arbitrum", "arbitrum-evmRepo"],
  ["arbitrum-sepolia", "arbitrum-evmRepo"],
  ["moonbeam", "moonbeam-evmRepo"],
  ["polygon", "polygon-evmRepo"],
  ["ethereum-holesky", "evmRepo"],
  ["scroll", "evmRepo"],
  ["polygon-sepolia", "polygon-evmRepo"],
  ["blast", "evmRepo"],
  ["mantle", "evmRepo"],
  ["xlayer", "evmRepo"],
]);

const POOL_STRATEGY = "weighted";

export class RepositoriesBuilder {
  private repositories = new Map();
  private snsClient?: SNSClient;
  private cfg: Config;

  constructor(cfg: Config) {
    this.cfg = cfg;
    this.build();
  }

  private build(): void {
    this.snsClient = this.createSnsClient();
    this.repositories.set("sns", new SnsEventRepository(this.snsClient, this.cfg.sns));

    this.cfg.metadata?.dir &&
      this.repositories.set("metadata", new FileMetadataRepository(this.cfg.metadata.dir));

    this.repositories.set("metrics", new PromStatRepository());

    this.cfg.enabledPlatforms.forEach((chain) => {
      this.buildWormchainRepository(chain);
      this.buildSolanaRepository(chain);
      this.buildAptosRepository(chain);
      this.buildEvmRepository(chain);
      this.buildSuiRepository(chain);
    });

    this.repositories.set(
      "jobs",
      new StaticJobRepository(
        this.cfg.environment,
        this.cfg.jobs.dir,
        this.cfg.dryRun,
        (chain: string) => this.getEvmBlockRepository(chain),
        {
          metadataRepo: this.getMetadataRepository(),
          statsRepo: this.getStatsRepository(),
          snsRepo: this.getSnsEventRepository(),
          solanaSlotRepo: this.getSolanaSlotRepository(),
          suiRepo: this.getSuiRepository(),
          aptosRepo: this.getAptosRepository(),
          wormchainRepo: this.getWormchainRepository(),
        }
      )
    );
  }

  public getEvmBlockRepository(chain: string): EvmJsonRPCBlockRepository {
    const instanceRepoName = EVM_CHAINS.get(chain);
    if (!instanceRepoName)
      throw new Error(`[RepositoriesBuilder] Chain ${chain.toLocaleUpperCase()} not supported`);
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

  public getSuiRepository(): SuiRepository {
    return this.getRepo("sui-repo");
  }

  public getAptosRepository(): AptosRepository {
    return this.getRepo("aptos-repo");
  }

  public getWormchainRepository(): WormchainRepository {
    return this.getRepo("wormchain-repo");
  }

  public close(): void {
    this.snsClient?.destroy();
  }

  private buildSolanaRepository(chain: string): void {
    if (chain == SOLANA_CHAIN) {
      const solanaProviderPool = providerPoolSupplier(
        this.cfg.chains[chain].rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) =>
          new InstrumentedConnection(rpcCfg.url, {
            commitment: rpcCfg.commitment || "confirmed",
          }),
        POOL_STRATEGY
      );

      const cfg = this.cfg.chains[chain];
      const solanaSlotRepository = new RateLimitedSolanaSlotRepository(
        new Web3SolanaSlotRepository(solanaProviderPool),
        cfg.rateLimit
      );
      this.repositories.set("solana-slotRepo", solanaSlotRepository);
    }
  }

  private buildEvmRepository(chain: string): void {
    if (chain == EVM_CHAIN) {
      const pools = this.createEvmProviderPools();
      const repoCfg: EvmJsonRPCBlockRepositoryCfg = {
        chains: this.cfg.chains,
        environment: this.cfg.environment,
      };

      const moonbeamRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new MoonbeamEvmJsonRPCBlockRepository(repoCfg, pools)
      );
      const arbitrumRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new ArbitrumEvmJsonRPCBlockRepository(repoCfg, pools, this.getMetadataRepository())
      );
      const polygonRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new PolygonJsonRPCBlockRepository(repoCfg, pools)
      );
      const bscRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new BscEvmJsonRPCBlockRepository(repoCfg, pools)
      );
      const evmRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new EvmJsonRPCBlockRepository(repoCfg, pools)
      );

      this.repositories.set("moonbeam-evmRepo", moonbeamRepository);
      this.repositories.set("arbitrum-evmRepo", arbitrumRepository);
      this.repositories.set("polygon-evmRepo", polygonRepository);
      this.repositories.set("bsc-evmRepo", bscRepository);
      this.repositories.set("evmRepo", evmRepository);
    }
  }

  private buildSuiRepository(chain: string): void {
    if (chain == SUI_CHAIN) {
      const suiProviderPool = providerPoolSupplier(
        this.cfg.chains[chain].rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) => new InstrumentedSuiClient(rpcCfg.url, 2000),
        POOL_STRATEGY
      );

      const suiRepository = new RateLimitedSuiJsonRPCBlockRepository(
        new SuiJsonRPCBlockRepository(suiProviderPool)
      );

      this.repositories.set("sui-repo", suiRepository);
    }
  }

  private buildAptosRepository(chain: string): void {
    if (chain == APTOS_CHAIN) {
      const pools = this.createDefaultProviderPools(chain);

      const aptosRepository = new RateLimitedAptosJsonRPCBlockRepository(
        new AptosJsonRPCBlockRepository(pools)
      );

      this.repositories.set("aptos-repo", aptosRepository);
    }
  }

  private buildWormchainRepository(chain: string): void {
    if (chain == WORMCHAIN_CHAIN) {
      const pools = this.createDefaultProviderPools(chain);

      const wormchainRepository = new RateLimitedWormchainJsonRPCBlockRepository(
        new WormchainJsonRPCBlockRepository(pools)
      );

      this.repositories.set("wormchain-repo", wormchainRepository);
    }
  }

  private getRepo(name: string): any {
    const repo = this.repositories.get(name);
    if (!repo)
      throw new Error(`[RepositoriesBuilder] Repository ${name.toLocaleLowerCase()} not supported`);

    return repo;
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

  private createEvmProviderPools(): ProviderPoolMap {
    let pools: ProviderPoolMap = {};
    for (const chain in this.cfg.chains) {
      const cfg = this.cfg.chains[chain];
      pools[chain] = providerPoolSupplier(
        cfg.rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) => this.createHttpClient(chain, rpcCfg.url),
        POOL_STRATEGY
      );
    }
    return pools;
  }

  private createDefaultProviderPools(chain: string) {
    const cfg = this.cfg.chains[chain];
    const pools = providerPoolSupplier(
      cfg.rpcs.map((url) => ({ url })),
      (rpcCfg: RpcConfig) => this.createHttpClient(chain, rpcCfg.url),
      POOL_STRATEGY
    );
    return pools;
  }

  private createHttpClient(chain: string, url: string): InstrumentedHttpProvider {
    return new InstrumentedHttpProvider({
      chain,
      url,
      retries: 3,
      timeout: 1_0000,
      initialDelay: 1_000,
      maxDelay: 30_000,
    });
  }
}
