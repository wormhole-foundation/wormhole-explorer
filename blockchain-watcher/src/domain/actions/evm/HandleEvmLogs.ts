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
  target: (parsed: T[], chain: string) => Promise<void>;
  statsRepo: StatRepository;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmLog, args: ReadonlyArray<any>) => T,
    target: (parsed: T[], chain: string) => Promise<void>,
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
          this.cfg.filter.addresses.includes(log.address.toLowerCase()) &&
          this.cfg.filter.topics.includes(log.topics[0].toLowerCase())
      )
      .map((log) => {
        const iface = new ethers.utils.Interface([this.cfg.abi]);
        const parsedLog = iface.parseLog(log);
        const logMap = this.mapper(log, parsedLog.args);
        this.report();
        return logMap;
      });

    await this.target(mappedItems, this.cfg.chain);
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
      filter: {
        addresses: cfg.filter.addresses.map((addr) => addr.toLowerCase()),
        topics: cfg.filter.topics.map((topic) => topic.toLowerCase()),
      },
      metricName: cfg.metricName,
      commitment: cfg.commitment,
      chainId: cfg.chainId,
      chain: cfg.chain,
      abi: cfg.abi,
      id: cfg.id,
    };
  }
}
