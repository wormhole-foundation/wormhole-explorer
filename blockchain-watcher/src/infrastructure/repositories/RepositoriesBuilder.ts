import { InstrumentedSuiClient, ProviderPool, RpcConfig } from "@xlabs/rpc-pool";
import { RateLimitedWormchainJsonRPCBlockRepository } from "./wormchain/RateLimitedWormchainJsonRPCBlockRepository";
import { RateLimitedAlgorandJsonRPCBlockRepository } from "./algorand/RateLimitedAlgorandJsonRPCBlockRepository";
import { RateLimitedCosmosJsonRPCBlockRepository } from "./cosmos/RateLimitedCosmosJsonRPCBlockRepository";
import { RateLimitedAptosJsonRPCBlockRepository } from "./aptos/RateLimitedAptosJsonRPCBlockRepository";
import { RateLimitedNearJsonRPCBlockRepository } from "./near/RateLimitedNearJsonRPCBlockRepository";
import { RateLimitedEvmJsonRPCBlockRepository } from "./evm/RateLimitedEvmJsonRPCBlockRepository";
import { RateLimitedSuiJsonRPCBlockRepository } from "./sui/RateLimitedSuiJsonRPCBlockRepository";
import { WormchainJsonRPCBlockRepository } from "./wormchain/WormchainJsonRPCBlockRepository";
import { AlgorandJsonRPCBlockRepository } from "./algorand/AlgorandJsonRPCBlockRepository";
import { InstrumentedConnectionWrapper } from "../rpc/http/InstrumentedConnectionWrapper";
import { CosmosJsonRPCBlockRepository } from "./cosmos/CosmosJsonRPCBlockRepository";
import { extendedProviderPoolSupplier } from "../rpc/http/ProviderPoolDecorator";
import { AptosJsonRPCBlockRepository } from "./aptos/AptosJsonRPCBlockRepository";
import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { NearJsonRPCBlockRepository } from "./near/NearJsonRPCBlockRepository";
import { InstrumentedHttpProvider } from "../rpc/http/InstrumentedHttpProvider";
import { ChainRPCConfig, Config } from "../config";
import { InfluxEventRepository } from "./target/InfluxEventRepository";
import { InfluxDB } from "@influxdata/influxdb-client";
import {
  WormchainRepository,
  AlgorandRepository,
  CosmosRepository,
  AptosRepository,
  NearRepository,
  JobRepository,
  SuiRepository,
} from "../../domain/repositories";
import {
  MoonbeamEvmJsonRPCBlockRepository,
  ArbitrumEvmJsonRPCBlockRepository,
  RateLimitedSolanaSlotRepository,
  PolygonJsonRPCBlockRepository,
  BscEvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepository,
  SuiJsonRPCBlockRepository,
  Web3SolanaSlotRepository,
  FileMetadataRepository,
  StaticJobRepository,
  PromStatRepository,
  SnsEventRepository,
} from ".";

const WORMCHAIN_CHAIN = "wormchain";
const ALGORAND_CHAIN = "algorand";
const SOLANA_CHAIN = "solana";
const COSMOS_CHAIN = "cosmos";
const APTOS_CHAIN = "aptos";
const NEAR_CHAIN = "near";
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
  ["snaxchain", "evmRepo"],
]);

const POOL_STRATEGY = "healthy";

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

    this.cfg.influx &&
      this.repositories.set(
        "infux",
        new InfluxEventRepository(
          new InfluxDB({
            url: this.cfg.influx.url,
            token: this.cfg.influx.token,
          }),
          this.cfg.influx
        )
      );

    this.cfg.metadata?.dir &&
      this.repositories.set("metadata", new FileMetadataRepository(this.cfg.metadata.dir));

    this.repositories.set("metrics", new PromStatRepository());

    const pools = this.createAllProvidersPool();

    this.cfg.enabledPlatforms.forEach((chain) => {
      // Set up all providers because we use various chains
      this.buildWormchainRepository(chain, pools);
      this.buildCosmosRepository(chain, pools);
      this.buildEvmRepository(chain, pools);
      // Set up the specific providers for the chain
      this.buildAlgorandRepository(chain);
      this.buildSolanaRepository(chain);
      this.buildAptosRepository(chain);
      this.buildSuiRepository(chain);
      this.buildNearRepository(chain);
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
          influxRepo: this.getInfluxEventRepository(),
          solanaSlotRepo: this.getSolanaSlotRepository(),
          suiRepo: this.getSuiRepository(),
          aptosRepo: this.getAptosRepository(),
          wormchainRepo: this.getWormchainRepository(),
          cosmosRepo: this.getCosmosRepository(),
          algorandRepo: this.getAlgorandRepository(),
          nearRepo: this.getNearRepository(),
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

  public getSnsEventRepository(): SnsEventRepository | undefined {
    try {
      const sns = this.getRepo("sns");
      return sns;
    } catch (e) {
      return;
    }
  }

  public getInfluxEventRepository(): InfluxEventRepository | undefined {
    try {
      const influx = this.getRepo("infux");
      return influx;
    } catch (e) {
      return;
    }
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

  public getCosmosRepository(): CosmosRepository {
    return this.getRepo("cosmos-repo");
  }

  public getAlgorandRepository(): AlgorandRepository {
    return this.getRepo("algorand-repo");
  }

  public getNearRepository(): NearRepository {
    return this.getRepo("near-repo");
  }

  public close(): void {
    this.snsClient?.destroy();
  }

  private buildSolanaRepository(chain: string): void {
    if (chain == SOLANA_CHAIN) {
      const cfg = this.cfg.chains[chain];

      const solanaProviderPool = extendedProviderPoolSupplier(
        this.cfg.chains[chain].rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) =>
          new InstrumentedConnectionWrapper(
            rpcCfg.url,
            {
              commitment: rpcCfg.commitment || "confirmed",
            },
            cfg.timeout ?? 1_000,
            SOLANA_CHAIN
          ),
        POOL_STRATEGY
      );

      const solanaSlotRepository = new RateLimitedSolanaSlotRepository(
        new Web3SolanaSlotRepository(solanaProviderPool),
        SOLANA_CHAIN,
        {
          period: cfg.rateLimit?.period ?? 10_000,
          limit: cfg.rateLimit?.limit ?? 1_000,
          interval: cfg.timeout ?? 1_000,
          attempts: cfg.retries ?? 10,
        }
      );
      this.repositories.set("solana-slotRepo", solanaSlotRepository);
    }
  }

  private buildEvmRepository(chain: string, pools: ProviderPoolMap): void {
    if (chain == EVM_CHAIN) {
      const repoCfg: JsonRPCBlockRepositoryCfg = {
        chains: this.cfg.chains,
        environment: this.cfg.environment,
      };

      const moonbeamRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new MoonbeamEvmJsonRPCBlockRepository(repoCfg, pools),
        "moonbeam"
      );
      const arbitrumRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new ArbitrumEvmJsonRPCBlockRepository(repoCfg, pools, this.getMetadataRepository()),
        "arbitrum"
      );
      const polygonRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new PolygonJsonRPCBlockRepository(repoCfg, pools),
        "polygon"
      );
      const bscRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new BscEvmJsonRPCBlockRepository(repoCfg, pools),
        "bsc"
      );
      const evmRepository = new RateLimitedEvmJsonRPCBlockRepository(
        new EvmJsonRPCBlockRepository(repoCfg, pools),
        "evm"
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
      const suiProviderPool = extendedProviderPoolSupplier(
        this.cfg.chains[chain].rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) => new InstrumentedSuiClient(rpcCfg.url, 2000),
        POOL_STRATEGY
      );

      const suiRepository = new RateLimitedSuiJsonRPCBlockRepository(
        new SuiJsonRPCBlockRepository(suiProviderPool),
        SUI_CHAIN
      );

      this.repositories.set("sui-repo", suiRepository);
    }
  }

  private buildAptosRepository(chain: string): void {
    if (chain == APTOS_CHAIN) {
      const pools = this.createDefaultProviderPools(chain);

      const aptosRepository = new RateLimitedAptosJsonRPCBlockRepository(
        new AptosJsonRPCBlockRepository(pools),
        APTOS_CHAIN
      );

      this.repositories.set("aptos-repo", aptosRepository);
    }
  }

  private buildCosmosRepository(chain: string, pools: ProviderPoolMap): void {
    if (chain == COSMOS_CHAIN) {
      const repoCfg: JsonRPCBlockRepositoryCfg = {
        chains: this.cfg.chains,
        environment: this.cfg.environment,
      };

      const CosmosRepository = new RateLimitedCosmosJsonRPCBlockRepository(
        new CosmosJsonRPCBlockRepository(repoCfg, pools),
        COSMOS_CHAIN
      );

      this.repositories.set("cosmos-repo", CosmosRepository);
    }
  }

  private buildWormchainRepository(chain: string, pools: ProviderPoolMap): void {
    if (chain == WORMCHAIN_CHAIN) {
      const repoCfg: JsonRPCBlockRepositoryCfg = {
        chains: this.cfg.chains,
        environment: this.cfg.environment,
      };

      const wormchainRepository = new RateLimitedWormchainJsonRPCBlockRepository(
        new WormchainJsonRPCBlockRepository(repoCfg, pools),
        WORMCHAIN_CHAIN
      );

      this.repositories.set("wormchain-repo", wormchainRepository);
    }
  }

  private buildAlgorandRepository(chain: string): void {
    if (chain == ALGORAND_CHAIN) {
      const algoIndexerRpcs = this.cfg.chains[chain].rpcs[1] as unknown as string[];
      const algoRpcs = this.cfg.chains[chain].rpcs[0] as unknown as string[];

      const algoIndexerPools = this.createDefaultProviderPools(chain, algoIndexerRpcs);
      const algoV2Pools = this.createDefaultProviderPools(chain, algoRpcs);

      const algorandRepository = new RateLimitedAlgorandJsonRPCBlockRepository(
        new AlgorandJsonRPCBlockRepository(algoV2Pools, algoIndexerPools),
        ALGORAND_CHAIN
      );

      this.repositories.set("algorand-repo", algorandRepository);
    }
  }

  private buildNearRepository(chain: string): void {
    if (chain == NEAR_CHAIN) {
      const pools = this.createDefaultProviderPools(chain);

      const aptosRepository = new RateLimitedNearJsonRPCBlockRepository(
        new NearJsonRPCBlockRepository(pools),
        NEAR_CHAIN
      );

      this.repositories.set("near-repo", aptosRepository);
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

  private createAllProvidersPool(): ProviderPoolMap {
    let pools: ProviderPoolMap = {};
    for (const chain in this.cfg.chains) {
      const cfg = this.cfg.chains[chain];
      pools[chain] = extendedProviderPoolSupplier(
        cfg.rpcs.map((url) => ({ url })),
        (rpcCfg: RpcConfig) => this.createHttpClient(chain, rpcCfg.url),
        POOL_STRATEGY
      );
    }
    return pools;
  }

  private createDefaultProviderPools(chain: string, rpcs?: string[]) {
    if (!rpcs) {
      rpcs = this.cfg.chains[chain].rpcs;
    }

    const pools = extendedProviderPoolSupplier(
      rpcs.map((url) => ({ url })),
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

export type JsonRPCBlockRepositoryCfg = {
  chains: Record<string, ChainRPCConfig>;
  environment: string;
};

export type ProviderPoolMap = Record<string, ProviderPool<InstrumentedHttpProvider>>;
