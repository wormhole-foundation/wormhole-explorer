import { InstrumentedConnection, InstrumentedSuiClient, RpcConfig } from "@xlabs/rpc-pool";
import { RateLimitedWormchainJsonRPCBlockRepository } from "./wormchain/RateLimitedWormchainJsonRPCBlockRepository";
import { RateLimitedAlgorandJsonRPCBlockRepository } from "./algorand/RateLimitedAlgorandJsonRPCBlockRepository";
import { RateLimitedAptosJsonRPCBlockRepository } from "./aptos/RateLimitedAptosJsonRPCBlockRepository";
import { RateLimitedEvmJsonRPCBlockRepository } from "./evm/RateLimitedEvmJsonRPCBlockRepository";
import { RateLimitedSeiJsonRPCBlockRepository } from "./sei/RateLimitedSeiJsonRPCBlockRepository";
import { RateLimitedSuiJsonRPCBlockRepository } from "./sui/RateLimitedSuiJsonRPCBlockRepository";
import { WormchainJsonRPCBlockRepository } from "./wormchain/WormchainJsonRPCBlockRepository";
import { AlgorandJsonRPCBlockRepository } from "./algorand/AlgorandJsonRPCBlockRepository";
import { AptosJsonRPCBlockRepository } from "./aptos/AptosJsonRPCBlockRepository";
import { SNSClient, SNSClientConfig } from "@aws-sdk/client-sns";
import { SeiJsonRPCBlockRepository } from "./sei/SeiJsonRPCBlockRepository";
import { InstrumentedHttpProvider } from "../rpc/http/InstrumentedHttpProvider";
import { Config } from "../config";
import {
  WormchainRepository,
  AlgorandRepository,
  AptosRepository,
  JobRepository,
  SuiRepository,
  SeiRepository,
} from "../../domain/repositories";
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
import {
  providerPoolSupplierDecorator,
  ProviderPoolDecorator,
} from "../rpc/http/ProviderPoolDecorator";

const WORMCHAIN_CHAIN = "wormchain";
const ALGORAND_CHAIN = "algorand";
const SOLANA_CHAIN = "solana";
const APTOS_CHAIN = "aptos";
const EVM_CHAIN = "evm";
const SUI_CHAIN = "sui";
const SEI_CHAIN = "sei";
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
      this.buildAlgorandRepository(chain);
      this.buildSolanaRepository(chain);
      this.buildAptosRepository(chain);
      this.buildEvmRepository(chain);
      this.buildSuiRepository(chain);
      this.buildSeiRepository(chain);
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
          seiRepo: this.getSeiRepository(),
          algorandRepo: this.getAlgorandRepository(),
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

  public getSeiRepository(): SeiRepository {
    return this.getRepo("sei-repo");
  }

  public getAlgorandRepository(): AlgorandRepository {
    return this.getRepo("algorand-repo");
  }

  public close(): void {
    this.snsClient?.destroy();
  }

  private buildSolanaRepository(chain: string): void {
    if (chain == SOLANA_CHAIN) {
      const solanaProviderPool = providerPoolSupplierDecorator(
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
      const suiProviderPool = providerPoolSupplierDecorator(
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

  private buildSeiRepository(chain: string): void {
    if (chain == SEI_CHAIN) {
      const pools = this.createDefaultProviderPools(chain);

      const seiRepository = new RateLimitedSeiJsonRPCBlockRepository(
        new SeiJsonRPCBlockRepository(pools)
      );

      this.repositories.set("sei-repo", seiRepository);
    }
  }

  private buildWormchainRepository(chain: string): void {
    if (chain == WORMCHAIN_CHAIN) {
      const injectivePools = this.createDefaultProviderPools("injective");
      const wormchainPools = this.createDefaultProviderPools("wormchain");
      const osmosisPools = this.createDefaultProviderPools("osmosis");
      const kujiraPools = this.createDefaultProviderPools("kujira");
      const evmosPools = this.createDefaultProviderPools("evmos");

      const cosmosPools: Map<number, ProviderPoolDecorator<InstrumentedHttpProvider>> = new Map([
        [19, injectivePools],
        [20, osmosisPools],
        [3104, wormchainPools],
        [4001, evmosPools],
        [4002, kujiraPools],
      ]);

      const wormchainRepository = new RateLimitedWormchainJsonRPCBlockRepository(
        new WormchainJsonRPCBlockRepository(cosmosPools)
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

      const seiRepository = new RateLimitedAlgorandJsonRPCBlockRepository(
        new AlgorandJsonRPCBlockRepository(algoV2Pools, algoIndexerPools)
      );

      this.repositories.set("algorand-repo", seiRepository);
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
      pools[chain] = providerPoolSupplierDecorator(
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

    const pools = providerPoolSupplierDecorator(
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
