import { HandleEvmLogsConfig } from "./types";
import { StatRepository } from "../../repositories";
import { ethers } from "ethers";
import { EvmLog } from "../../entities";

export class HandleEvmLogs<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (log: EvmLog, parsedArgs: ReadonlyArray<any>) => T;
  target: (parsed: T[]) => Promise<void>;
  statsRepo: StatRepository;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmLog, args: ReadonlyArray<any>) => T,
    target: (parsed: T[]) => Promise<void>,
    statsRepo: StatRepository
  ) {
    this.cfg = this.normalizeCfg(cfg);
    this.mapper = mapper;
    this.target = target;
    this.statsRepo = statsRepo;
  }

  public async handle(logs: EvmLog[]): Promise<T[]> {
    const mappedItems = logs
      .map((log) => {
        const iface = new ethers.utils.Interface([this.cfg.abi]);
        const parsedLog = iface.parseLog(log);
        const logMap = this.mapper(log, parsedLog.args);
        if (logMap) {
          this.report();
          return logMap;
        }
      })
      .filter((log) => log) as T[];

    await this.target(mappedItems);
    return mappedItems;
  }

  private report() {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.commitment,
    };
    this.statsRepo.count(this.cfg.metricName, labels);
  }

  private normalizeCfg(cfg: HandleEvmLogsConfig): HandleEvmLogsConfig {
    return {
      metricName: cfg.metricName,
      commitment: cfg.commitment,
      chainId: cfg.chainId,
      chain: cfg.chain,
      abi: cfg.abi,
      id: cfg.id,
      filters: cfg.filters.map((filter) => {
        return {
          addresses: filter.addresses.map((address) => address.toLowerCase()),
          topics: filter.topics.map((topic) => topic.toLowerCase()),
        };
      }),
    };
  }
}
