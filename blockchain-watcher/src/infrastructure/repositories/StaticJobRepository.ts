import { PollNear, PollNearConfig, PollNearConfigProps } from "../../domain/actions/near/PollNear";
import { FileMetadataRepository, SnsEventRepository } from "./index";
import { wormchainRedeemedTransactionFoundMapper } from "../mappers/wormchain/wormchainRedeemedTransactionFoundMapper";
import { algorandRedeemedTransactionFoundMapper } from "../mappers/algorand/algorandRedeemedTransactionFoundMapper";
import { JobDefinition, Handler, LogFoundEvent } from "../../domain/entities";
import { cosmosRedeemedTransactionFoundMapper } from "../mappers/cosmos/cosmosRedeemedTransactionFoundMapper";
import { aptosRedeemedTransactionFoundMapper } from "../mappers/aptos/aptosRedeemedTransactionFoundMapper";
import { nearRedeemedTransactionFoundMapper } from "../mappers/near/nearRedeemedTransactionFoundMapper";
import { wormchainLogMessagePublishedMapper } from "../mappers/wormchain/wormchainLogMessagePublishedMapper";
import { algorandLogMessagePublishedMapper } from "../mappers/algorand/algorandLogMessagePublishedMapper";
import { suiRedeemedTransactionFoundMapper } from "../mappers/sui/suiRedeemedTransactionFoundMapper";
import { solanaLogCircleMessageSentMapper } from "../mappers/solana/solanaLogCircleMessageSentMapper";
import { evmNttWormholeTransceiverMapper } from "../mappers/evm/evmNttWormholeTransceiverMapper";
import { cosmosLogMessagePublishedMapper } from "../mappers/cosmos/cosmosLogMessagePublishedMapper";
import { aptosLogMessagePublishedMapper } from "../mappers/aptos/aptosLogMessagePublishedMapper";
import { evmLogCircleMessageSentMapper } from "../mappers/evm/evmLogCircleMessageSentMapper";
import { evmNttAxelarTransceiverMapper } from "../mappers/evm/evmNttAxelarTransceiverMapper";
import { evmNttMessageAttestedToMapper } from "../mappers/evm/evmNttMessageAttestedToMapper";
import { evmNttTransferRedeemedMapper } from "../mappers/evm/evmNttTransferRedeemedMapper";
import { suiLogMessagePublishedMapper } from "../mappers/sui/suiLogMessagePublishedMapper";
import { HandleAlgorandTransactions } from "../../domain/actions/algorand/HandleAlgorandTransactions";
import { HandleSolanaTransactions } from "../../domain/actions/solana/HandleSolanaTransactions";
import { HandleCosmosTransactions } from "../../domain/actions/cosmos/HandleCosmosTransactions";
import { evmNttTransferSentMapper } from "../mappers/evm/evmNttTransferSentMapper";
import { evmProposalCreatedMapper } from "../mappers/evm/evmProposalCreatedMapper";
import { HandleAptosTransactions } from "../../domain/actions/aptos/HandleAptosTransactions";
import { HandleNearTransactions } from "../../domain/actions/near/HandleNearTransactions";
import { HandleWormchainRedeems } from "../../domain/actions/wormchain/HandleWormchainRedeems";
import { HandleEvmTransactions } from "../../domain/actions/evm/HandleEvmTransactions";
import { HandleSuiTransactions } from "../../domain/actions/sui/HandleSuiTransactions";
import { InfluxEventRepository } from "./target/InfluxEventRepository";
import { HandleWormchainLogs } from "../../domain/actions/wormchain/HandleWormchainLogs";
import { SqsEventRepository } from "./target/SqsEventRepository";
import { RunRPCHealthcheck } from "../../domain/actions/RunRPCHealthcheck";
import { RPCHealthcheck } from "../../domain/RPCHealthcheck/RPCHealthcheck";
import { Registry } from "../../domain/registry/Registry";
import log from "../log";
import {
  PollCosmosConfigProps,
  PollCosmosConfig,
  PollCosmos,
} from "../../domain/actions/cosmos/PollCosmos";
import {
  PollWormchainLogsConfigProps,
  PollWormchainLogsConfig,
  PollWormchain,
} from "../../domain/actions/wormchain/PollWormchain";
import {
  SolanaSlotRepository,
  WormchainRepository,
  AlgorandRepository,
  EvmBlockRepository,
  MetadataRepository,
  CosmosRepository,
  AptosRepository,
  NearRepository,
  StatRepository,
  SuiRepository,
  JobRepository,
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
import {
  PollAlgorandConfigProps,
  PollAlgorandConfig,
  PollAlgorand,
} from "../../domain/actions/algorand/PollAlgorand";

export class StaticJobRepository implements JobRepository {
  private fileRepo: FileMetadataRepository;
  private environment: string;
  private dryRun: boolean = false;
  private rpcHealthcheckInterval;
  private runPollingJob: Map<string, (jobDef: JobDefinition) => RunPollingJob> = new Map();
  private handlers: Map<string, (cfg: any, target: string, mapper: any) => Promise<Handler>> =
    new Map();
  private mappers: Map<string, any> = new Map();
  private targets: Map<string, () => Promise<(items: any[]) => Promise<void>>> = new Map();
  private repos: Repos;

  constructor(
    environment: string,
    path: string,
    dryRun: boolean,
    rpcHealthcheckInterval: number,
    repos: Repos
  ) {
    this.rpcHealthcheckInterval = rpcHealthcheckInterval;
    this.fileRepo = new FileMetadataRepository(path);
    this.environment = environment;
    this.dryRun = dryRun;
    this.repos = repos;
    this.fill();
  }

  async getJobDefinitions(): Promise<JobDefinition[]> {
    const persisted = await this.fileRepo.get("jobs");
    if (!persisted) {
      return Promise.resolve([]);
    }
    return persisted;
  }

  getPollingJob(jobDef: JobDefinition): RunPollingJob {
    const action = this.runPollingJob.get(jobDef.source.action);
    if (!action) {
      throw new Error(`Source ${jobDef.source.action} not found`);
    }
    return action(jobDef);
  }

  getRPCHealthcheck(jobsDef: JobDefinition[]): RunRPCHealthcheck {
    const rpcHealthcheck = (jobsDef: JobDefinition[]) =>
      new RPCHealthcheck(
        this.repos.statsRepo,
        this.repos.metadataRepo,
        jobsDef.map((jobDef) => ({
          repository:
            jobDef.source.repository == "evmRepo"
              ? this.repos.evmRepo(jobDef.chain)
              : this.repos[jobDef.source.repository as keyof Repos],
          environment: jobDef.source.config.environment,
          commitment: jobDef.source.config.commitment,
          interval: jobDef.source.config.interval,
          chainId: jobDef.source.config.chainId,
          chain: jobDef.chain,
          id: jobDef.id,
        })),
        this.rpcHealthcheckInterval
      );
    return rpcHealthcheck(jobsDef);
  }

  async getRegistry(jobDef: JobDefinition): Promise<any> {
    const registry = (jobDef: JobDefinition) =>
      new Registry(this.repos.statsRepo, this.repos.metadataRepo, this.repos.sqsRepo!, {
        environment: jobDef.source.config.environment,
        commitment: jobDef.source.config.commitment,
        interval: jobDef.source.config.interval,
        chainId: jobDef.source.config.chainId,
        chain: jobDef.chain,
        id: jobDef.id,
      });
    return registry(jobDef);
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
    const {
      evmRepo,
      metadataRepo,
      statsRepo,
      solanaSlotRepo,
      suiRepo,
      aptosRepo,
      wormchainRepo,
      cosmosRepo,
      algorandRepo,
      nearRepo,
    } = this.repos;

    const pollEvm = (jobDef: JobDefinition) =>
      new PollEvm(
        evmRepo(jobDef.source.config.chain),
        metadataRepo,
        statsRepo,
        new PollEvmLogsConfig({
          ...(jobDef.source.config as PollEvmLogsConfigProps),
          id: jobDef.id,
          environment: this.environment,
        }),
        jobDef.source.records
      );
    const pollSolanaTransactions = (jobDef: JobDefinition) =>
      new PollSolanaTransactions(metadataRepo, solanaSlotRepo, statsRepo, {
        ...(jobDef.source.config as PollSolanaTransactionsConfig),
        id: jobDef.id,
      });
    const pollSuiTransactions = (jobDef: JobDefinition) =>
      new PollSuiTransactions(
        new PollSuiTransactionsConfig({ ...jobDef.source.config, id: jobDef.id }),
        statsRepo,
        metadataRepo,
        suiRepo
      );
    const pollAptos = (jobDef: JobDefinition) =>
      new PollAptos(
        new PollAptosTransactionsConfig({
          ...(jobDef.source.config as PollAptosTransactionsConfigProps),
          id: jobDef.id,
          environment: this.environment,
        }),
        statsRepo,
        metadataRepo,
        aptosRepo,
        jobDef.source.records
      );
    const pollWormchain = (jobDef: JobDefinition) =>
      new PollWormchain(
        wormchainRepo,
        metadataRepo,
        statsRepo,
        new PollWormchainLogsConfig({
          ...(jobDef.source.config as PollWormchainLogsConfigProps),
          id: jobDef.id,
        }),
        jobDef.source.records
      );
    const pollComsos = (jobDef: JobDefinition) =>
      new PollCosmos(
        cosmosRepo,
        metadataRepo,
        statsRepo,
        new PollCosmosConfig({
          ...(jobDef.source.config as PollCosmosConfigProps),
          id: jobDef.id,
        })
      );
    const pollAlgorand = (jobDef: JobDefinition) =>
      new PollAlgorand(
        algorandRepo,
        metadataRepo,
        statsRepo,
        new PollAlgorandConfig({
          ...(jobDef.source.config as PollAlgorandConfigProps),
          id: jobDef.id,
        })
      );
    const pollNear = (jobDef: JobDefinition) =>
      new PollNear(
        nearRepo,
        metadataRepo,
        statsRepo,
        new PollNearConfig({
          ...(jobDef.source.config as PollNearConfigProps),
          id: jobDef.id,
        })
      );

    this.runPollingJob.set("PollEvm", pollEvm);
    this.runPollingJob.set("PollSolanaTransactions", pollSolanaTransactions);
    this.runPollingJob.set("PollSuiTransactions", pollSuiTransactions);
    this.runPollingJob.set("PollAptos", pollAptos);
    this.runPollingJob.set("PollWormchain", pollWormchain);
    this.runPollingJob.set("PollCosmos", pollComsos);
    this.runPollingJob.set("PollAlgorand", pollAlgorand);
    this.runPollingJob.set("PollNear", pollNear);
  }

  private loadMappers(): void {
    this.mappers.set("evmLogMessagePublishedMapper", evmLogMessagePublishedMapper);
    this.mappers.set("evmLogCircleMessageSentMapper", evmLogCircleMessageSentMapper);
    this.mappers.set("evmRedeemedTransactionFoundMapper", evmRedeemedTransactionFoundMapper);
    this.mappers.set("evmNttTransferSentMapper", evmNttTransferSentMapper);
    this.mappers.set("evmNttAxelarTransceiverMapper", evmNttAxelarTransceiverMapper);
    this.mappers.set("evmNttWormholeTransceiverMapper", evmNttWormholeTransceiverMapper);
    this.mappers.set("evmNttMessageAttestedToMapper", evmNttMessageAttestedToMapper);
    this.mappers.set("evmNttTransferRedeemedMapper", evmNttTransferRedeemedMapper);
    this.mappers.set("evmProposalCreatedMapper", evmProposalCreatedMapper);
    this.mappers.set("solanaLogMessagePublishedMapper", solanaLogMessagePublishedMapper);
    this.mappers.set("solanaTransferRedeemedMapper", solanaTransferRedeemedMapper);
    this.mappers.set("solanaLogCircleMessageSentMapper", solanaLogCircleMessageSentMapper);
    this.mappers.set("suiLogMessagePublishedMapper", suiLogMessagePublishedMapper);
    this.mappers.set("suiRedeemedTransactionFoundMapper", suiRedeemedTransactionFoundMapper);
    this.mappers.set("aptosLogMessagePublishedMapper", aptosLogMessagePublishedMapper);
    this.mappers.set("aptosRedeemedTransactionFoundMapper", aptosRedeemedTransactionFoundMapper);
    this.mappers.set("wormchainLogMessagePublishedMapper", wormchainLogMessagePublishedMapper);
    this.mappers.set("cosmosRedeemedTransactionFoundMapper", cosmosRedeemedTransactionFoundMapper);
    this.mappers.set(
      "algorandRedeemedTransactionFoundMapper",
      algorandRedeemedTransactionFoundMapper
    );
    this.mappers.set("algorandLogMessagePublishedMapper", algorandLogMessagePublishedMapper);
    this.mappers.set(
      "wormchainRedeemedTransactionFoundMapper",
      wormchainRedeemedTransactionFoundMapper
    );
    this.mappers.set("cosmosLogMessagePublishedMapper", cosmosLogMessagePublishedMapper);
    this.mappers.set("nearRedeemedTransactionFoundMapper", nearRedeemedTransactionFoundMapper);
  }

  private loadTargets(): void {
    const { snsRepo, influxRepo } = this.repos;

    const snsTarget = () => snsRepo!.asTarget();
    const influxTarget = () => influxRepo!.asTarget();
    const dummyTarget = async () => async (events: any[]) => {
      log.info(`[target dummy] Got ${events.length} events`);
    };

    snsRepo && this.targets.set("sns", snsTarget);
    influxRepo && this.targets.set("influx", influxTarget);
    this.targets.set("dummy", dummyTarget);
  }

  private loadHandlers(): void {
    const { statsRepo } = this.repos;

    const handleEvmLogs = async (config: any, target: string, mapper: any) => {
      const instance = new HandleEvmLogs<LogFoundEvent<any>>(
        config,
        mapper,
        await this.targets.get(this.dryRun ? "dummy" : target)!(),
        statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleEvmTransactions = async (config: any, target: string, mapper: any) => {
      const instance = new HandleEvmTransactions<LogFoundEvent<any>>(
        config,
        mapper,
        await this.targets.get(this.dryRun ? "dummy" : target)!(),
        statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleSolanaTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleSolanaTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleSuiTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleSuiTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };
    const handleAptosTx = async (config: any, target: string, mapper: any) => {
      const instance = new HandleAptosTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleWormchainLogs = async (config: any, target: string, mapper: any) => {
      const instance = new HandleWormchainLogs(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleWormchainRedeems = async (config: any, target: string, mapper: any) => {
      const instance = new HandleWormchainRedeems(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleCosmosTransactions = async (config: any, target: string, mapper: any) => {
      const instance = new HandleCosmosTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleAlgorandTransactions = async (config: any, target: string, mapper: any) => {
      const instance = new HandleAlgorandTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    const handleNearTransactions = async (config: any, target: string, mapper: any) => {
      const instance = new HandleNearTransactions(
        config,
        mapper,
        await this.getTarget(target),
        statsRepo
      );
      return instance.handle.bind(instance);
    };

    this.handlers.set("HandleEvmLogs", handleEvmLogs);
    this.handlers.set("HandleEvmTransactions", handleEvmTransactions);
    this.handlers.set("HandleSolanaTransactions", handleSolanaTx);
    this.handlers.set("HandleSuiTransactions", handleSuiTx);
    this.handlers.set("HandleAptosTransactions", handleAptosTx);
    this.handlers.set("HandleWormchainLogs", handleWormchainLogs);
    this.handlers.set("HandleWormchainRedeems", handleWormchainRedeems);
    this.handlers.set("HandleCosmosTransactions", handleCosmosTransactions);
    this.handlers.set("HandleAlgorandTransactions", handleAlgorandTransactions);
    this.handlers.set("HandleNearTransactions", handleNearTransactions);
  }

  private async getTarget(target: string): Promise<(items: any[]) => Promise<void>> {
    const maybeTarget = this.targets.get(this.dryRun ? "dummy" : target);
    if (!maybeTarget) {
      throw new Error(`Target ${target} not found`);
    }
    return maybeTarget();
  }
}

export type Repos = {
  evmRepo: (chain: string) => EvmBlockRepository;
  metadataRepo: MetadataRepository<any>;
  statsRepo: StatRepository;
  snsRepo?: SnsEventRepository;
  sqsRepo?: SqsEventRepository;
  influxRepo?: InfluxEventRepository;
  solanaSlotRepo: SolanaSlotRepository;
  suiRepo: SuiRepository;
  aptosRepo: AptosRepository;
  wormchainRepo: WormchainRepository;
  cosmosRepo: CosmosRepository;
  algorandRepo: AlgorandRepository;
  nearRepo: NearRepository;
};
