import { FileMetadataRepository, SnsEventRepository } from "./index";
import { JobDefinition, Handler, LogFoundEvent } from "../../domain/entities";
import { aptosRedeemedTransactionFoundMapper } from "../mappers/aptos/aptosRedeemedTransactionFoundMapper";
import { wormchainLogMessagePublishedMapper } from "../mappers/wormchain/wormchainLogMessagePublishedMapper";
import { suiRedeemedTransactionFoundMapper } from "../mappers/sui/suiRedeemedTransactionFoundMapper";
import { aptosLogMessagePublishedMapper } from "../mappers/aptos/aptosLogMessagePublishedMapper";
import { suiLogMessagePublishedMapper } from "../mappers/sui/suiLogMessagePublishedMapper";
import { HandleSolanaTransactions } from "../../domain/actions/solana/HandleSolanaTransactions";
import { HandleAptosTransactions } from "../../domain/actions/aptos/HandleAptosTransactions";
import { HandleEvmTransactions } from "../../domain/actions/evm/HandleEvmTransactions";
import { HandleSuiTransactions } from "../../domain/actions/sui/HandleSuiTransactions";
import { HandleWormchainLogs } from "../../domain/actions/wormchain/HandleWormchainLogs";
import log from "../log";
import {
  PollWormchainLogsConfigProps,
  PollWormchainLogsConfig,
  PollWormchain,
} from "../../domain/actions/wormchain/PollWormchain";
import {
  SolanaSlotRepository,
  EvmBlockRepository,
  MetadataRepository,
  AptosRepository,
  StatRepository,
  JobRepository,
  SuiRepository,
  WormchainRepository,
} from "../../domain/repositories";
import {
  PollSolanaTransactionsConfig,
  PollSolanaTransactions,
  PollEvmLogsConfigProps,
  PollEvmLogsConfig,
  RunPollingJob,
  HandleEvmLogs,
  PollEvm,
} from "../../domain/actions";
import {
  evmRedeemedTransactionFoundMapper,
  solanaLogMessagePublishedMapper,
  solanaTransferRedeemedMapper,
  evmLogMessagePublishedMapper,
} from "../mappers";
import {
  PollSuiTransactionsConfig,
  PollSuiTransactions,
} from "../../domain/actions/sui/PollSuiTransactions";
import {
  PollAptosTransactionsConfigProps,
  PollAptosTransactionsConfig,
  PollAptos,
} from "../../domain/actions/aptos/PollAptos";

export class StaticJobRepository implements JobRepository {
  private fileRepo: FileMetadataRepository;
  private environment: string;
  private dryRun: boolean = false;
  private sources: Map<string, (def: JobDefinition) => RunPollingJob> = new Map();
  private handlers: Map<string, (cfg: any, target: string, mapper: any) => Promise<Handler>> =
    new Map();
  private mappers: Map<string, any> = new Map();
  private targets: Map<string, () => Promise<(items: any[]) => Promise<void>>> = new Map();
  private blockRepoProvider: (chain: string) => EvmBlockRepository;
  private metadataRepo: MetadataRepository<any>;
  private statsRepo: StatRepository;
  private snsRepo: SnsEventRepository;
  private solanaSlotRepo: SolanaSlotRepository;
  private suiRepo: SuiRepository;
  private aptosRepo: AptosRepository;
  private wormchainRepo: WormchainRepository;

  constructor(
    environment: string,
    path: string,
    dryRun: boolean,
    blockRepoProvider: (chain: string) => EvmBlockRepository,
    repos: {
      metadataRepo: MetadataRepository<any>;
      statsRepo: StatRepository;
      snsRepo: SnsEventRepository;
      solanaSlotRepo: SolanaSlotRepository;
      suiRepo: SuiRepository;
      aptosRepo: AptosRepository;
      wormchainRepo: WormchainRepository;
    }
  ) {
    this.fileRepo = new FileMetadataRepository(path);
    this.blockRepoProvider = blockRepoProvider;
    this.metadataRepo = repos.metadataRepo;
    this.statsRepo = repos.statsRepo;
    this.snsRepo = repos.snsRepo;
    this.solanaSlotRepo = repos.solanaSlotRepo;
    this.suiRepo = repos.suiRepo;
    this.aptosRepo = repos.aptosRepo;
    this.wormchainRepo = repos.wormchainRepo;
    this.environment = environment;
    this.dryRun = dryRun;
    this.fill();
  }

  async getJobDefinitions(): Promise<JobDefinition[]> {
    const persisted = await this.fileRepo.get("jobs");
    if (!persisted) {
      return Promise.resolve([]);
    }
    return persisted;
  }

  getSource(jobDef: JobDefinition): RunPollingJob {
    const src = this.sources.get(jobDef.source.action);
    if (!src) {
      throw new Error(`Source ${jobDef.source.action} not found`);
    }
    return src(jobDef);
  }

  async getHandlers(jobDef: JobDefinition): Promise<Handler[]> {
    const result: Handler[] = [];
    for (const handler of jobDef.handlers) {
      const maybeHandler = this.handlers.get(handler.action);
      if (!maybeHandler) {
        throw new Error(`Handler ${handler.action} not found`);
      }
      const mapper = this.mappers.get(handler.mapper);
      if (!mapper) {
        throw new Error(`Handler ${handler.mapper} not found`);
      }

      const config = {
        ...(handler.config as any),
        commitment: jobDef.source.config.commitment,
        filters: jobDef.source.config.filters,
        chainId: jobDef.chainId,
        chain: jobDef.chain,
        id: jobDef.id,
      };
      result.push((await maybeHandler(config, handler.target, mapper)).bind(maybeHandler));
    }
    return result;
  }

  /**
   * Fill all resources that applications needs to work
   * Resources are: actions, mappers, targets and handlers
   */
  private fill() {
    this.loadActions();
    this.loadMappers();
    this.loadTargets();
    this.loadHandlers();
  }

  private loadActions(): void {
    const pollEvm = (jobDef: JobDefinition) =>
      new PollEvm(
        this.blockRepoProvider(jobDef.source.config.chain),
        this.metadataRepo,
        this.statsRepo,
        new PollEvmLogsConfig({
          ...(jobDef.source.config as PollEvmLogsConfigProps),
          id: jobDef.id,
          environment: this.environment,
        }),
        jobDef.source.records
      );
    const pollSolanaTransactions = (jobDef: JobDefinition) =>
      new PollSolanaTransactions(this.metadataRepo, this.solanaSlotRepo, this.statsRepo, {
        ...(jobDef.source.config as PollSolanaTransactionsConfig),
        id: jobDef.id,
      });
    const pollSuiTransactions = (jobDef: JobDefinition) =>
      new PollSuiTransactions(
        new PollSuiTransactionsConfig({ ...jobDef.source.config, id: jobDef.id }),
        this.statsRepo,
        this.metadataRepo,
        this.suiRepo
      );
    const pollAptos = (jobDef: JobDefinition) =>
      new PollAptos(
        new PollAptosTransactionsConfig({
          ...(jobDef.source.config as PollAptosTransactionsConfigProps),
          id: jobDef.id,
          environment: this.environment,
        }),
        this.statsRepo,
        this.metadataRepo,
        this.aptosRepo,
        jobDef.source.records
      );
    const pollWormchain = (jobDef: JobDefinition) =>
      new PollWormchain(
        this.wormchainRepo,
        this.metadataRepo,
        this.statsRepo,
        new PollWormchainLogsConfig({
          ...(jobDef.source.config as PollWormchainLogsConfigProps),
          id: jobDef.id,
        }),
        jobDef.source.records
      );

    this.sources.set("PollEvm", pollEvm);
    this.sources.set("PollSolanaTransactions", pollSolanaTransactions);
    this.sources.set("PollSuiTransactions", pollSuiTransactions);
    this.sources.set("PollAptos", pollAptos);
    this.sources.set("PollWormchain", pollWormchain);
  }

  private loadMappers(): void {
    this.mappers.set("evmLogMessagePublishedMapper", evmLogMessagePublishedMapper);
    this.mappers.set("evmRedeemedTransactionFoundMapper", evmRedeemedTransactionFoundMapper);
    this.mappers.set("solanaLogMessagePublishedMapper", solanaLogMessagePublishedMapper);
    this.mappers.set("solanaTransferRedeemedMapper", solanaTransferRedeemedMapper);
    this.mappers.set("suiLogMessagePublishedMapper", suiLogMessagePublishedMapper);
    this.mappers.set("suiRedeemedTransactionFoundMapper", suiRedeemedTransactionFoundMapper);
    this.mappers.set("aptosLogMessagePublishedMapper", aptosLogMessagePublishedMapper);
    this.mappers.set("aptosRedeemedTransactionFoundMapper", aptosRedeemedTransactionFoundMapper);
    this.mappers.set("wormchainLogMessagePublishedMapper", wormchainLogMessagePublishedMapper);
  }

  private loadTargets(): void {
    const snsTarget = () => this.snsRepo.asTarget();
    const dummyTarget = async () => async (events: any[]) => {
      log.info(`[target dummy] Got ${events.length} events`);
    };

    this.targets.set("sns", snsTarget);
    this.targets.set("dummy", dummyTarget);
  }

  private loadHandlers(): void {
    const handleEvmLogs = async (config: any, target: string, mapper: any) => {
      const instance = new HandleEvmLogs<LogFoundEvent<any>>(
        config,
        mapper,
        await this.targets.get(this.dryRun ? "dummy" : target)!(),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleEvmTransactions = async (config: any, target: string, mapper: any) => {
      const instance = new HandleEvmTransactions<LogFoundEvent<any>>(
        config,
        mapper,
        await this.targets.get(this.dryRun ? "dummy" : target)!(),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleSolanaTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleSolanaTransactions(
        config,
        mapper,
        await this.getTarget(target),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleSuiTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleSuiTransactions(
        config,
        mapper,
        await this.getTarget(target),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleAptosTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleAptosTransactions(
        config,
        mapper,
        await this.getTarget(target),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleWormchainLogs = async (config: any, target: string, mapper: any) => {
      const instance = new HandleWormchainLogs(
        config,
        mapper,
        await this.getTarget(target),
        this.statsRepo
      );
      return instance.handle.bind(instance);
    };

    this.handlers.set("HandleEvmLogs", handleEvmLogs);
    this.handlers.set("HandleEvmTransactions", handleEvmTransactions);
    this.handlers.set("HandleSolanaTransactions", handleSolanaTx);
    this.handlers.set("HandleSuiTransactions", handleSuiTx);
    this.handlers.set("HandleAptosTransactions", handleAptosTx);
    this.handlers.set("HandleWormchainLogs", handleWormchainLogs);
  }

  private async getTarget(target: string): Promise<(items: any[]) => Promise<void>> {
    const maybeTarget = this.targets.get(this.dryRun ? "dummy" : target);
    if (!maybeTarget) {
      throw new Error(`Target ${target} not found`);
    }
    return maybeTarget();
  }
}
