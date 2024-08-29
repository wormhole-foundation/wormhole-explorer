import { HandleEvmLogsConfig } from "./types";
import { StatRepository } from "../../repositories";
import { ethers } from "ethers";
import { EvmLog } from "../../entities";

/**
 * Handling means mapping and forward to a given target.
 * As of today, only one type of event can be handled per each instance.
 */
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
      .filter(
        (log) =>
          this.cfg.filters[0].addresses.includes(log.address.toLowerCase()) &&
          this.cfg.filters[0].topics.includes(log.topics[0].toLowerCase())
      )
      .map((log) => {
        const abi = this.cfg.abis.find((abi) => abi.topic === log.topics[0]) ?? this.cfg.abis[0];
        const iface = new ethers.utils.Interface([abi!.abi]);
        const parsedLog = iface.parseLog(log);
        const logMap = this.mapper(log, parsedLog.args);
        this.report();
        return logMap;
      });

    await this.target(mappedItems);
    // TODO: return a result specifying failures if any
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
      environment: cfg.environment,
      metricName: cfg.metricName,
      commitment: cfg.commitment,
      chainId: cfg.chainId,
      chain: cfg.chain,
      abis: cfg.abis,
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
