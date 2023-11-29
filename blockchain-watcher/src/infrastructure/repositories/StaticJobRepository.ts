import {
  HandleEvmLogs,
  PollEvmLogs,
  PollEvmLogsConfig,
  PollEvmLogsConfigProps,
  RunPollingJob,
} from "../../domain/actions";
import { JobDefinition, Handler, LogFoundEvent } from "../../domain/entities";
import {
  EvmBlockRepository,
  JobRepository,
  MetadataRepository,
  StatRepository,
} from "../../domain/repositories";
import { FileMetadataRepo, SnsEventRepository } from "./index";
import { evmLogMessagePublishedMapper } from "../mappers/evmLogMessagePublishedMapper";
import log from "../log";

export class StaticJobRepository implements JobRepository {
  private fileRepo: FileMetadataRepo;
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

  constructor(
    path: string,
    dryRun: boolean,
    blockRepoProvider: (chain: string) => EvmBlockRepository,
    repos: {
      metadataRepo: MetadataRepository<any>;
      statsRepo: StatRepository;
      snsRepo: SnsEventRepository;
    }
  ) {
    this.fileRepo = new FileMetadataRepo(path);
    this.blockRepoProvider = blockRepoProvider;
    this.metadataRepo = repos.metadataRepo;
    this.statsRepo = repos.statsRepo;
    this.snsRepo = repos.snsRepo;
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
        throw new Error(`Handler ${handler.action} not found`);
      }
      result.push((await maybeHandler(handler.config, handler.target, mapper)).bind(maybeHandler));
    }

    return result;
  }

  private fill() {
    const pollEvmLogs = (jobDef: JobDefinition) =>
      new PollEvmLogs(
        this.blockRepoProvider(jobDef.source.config.chain),
        this.metadataRepo,
        this.statsRepo,
        new PollEvmLogsConfig({
          ...(jobDef.source.config as PollEvmLogsConfigProps),
          id: jobDef.id,
        })
      );
    this.sources.set("PollEvmLogs", pollEvmLogs);

    this.mappers.set("evmLogMessagePublishedMapper", evmLogMessagePublishedMapper);

    const snsTarget = () => this.snsRepo.asTarget();
    const dummyTarget = async () => async (events: any[]) => {
      log.info(`Got ${events.length} events`);
    };
    this.targets.set("sns", snsTarget);
    this.targets.set("dummy", dummyTarget);

    const handleEvmLogs = async (config: any, target: string, mapper: any) => {
      const instance = new HandleEvmLogs<LogFoundEvent<any>>(
        config,
        mapper,
        await this.targets.get(this.dryRun ? "dummy" : target)!()
      );

      return instance.handle.bind(instance);
    };

    this.handlers.set("HandleEvmLogs", handleEvmLogs);
  }
}
