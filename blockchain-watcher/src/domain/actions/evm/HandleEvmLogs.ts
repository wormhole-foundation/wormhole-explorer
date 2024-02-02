import { EvmLog, EvmTopicFilter } from "../../entities";
import { StatRepository } from "../../repositories";
import { ethers } from "ethers";

/**
 * Handling means mapping and forward to a given target.
 * As of today, only one type of event can be handled per each instance.
 */
export class HandleEvmLogs<T> {
  cfg: HandleEvmConfig;
  mapper: (log: EvmLog, parsedArgs: ReadonlyArray<any>) => T;
  target: (parsed: T[]) => Promise<void>;
  statsRepo: StatRepository;

  constructor(
    cfg: HandleEvmConfig,
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
    this.statsRepo.count("process_source_event", labels);
  }

  private normalizeCfg(cfg: HandleEvmConfig): HandleEvmConfig {
    return {
      filter: {
        addresses: cfg.filter.addresses.map((addr) => addr.toLowerCase()),
        topics: cfg.filter.topics.map((topic) => topic.toLowerCase()),
      },
      commitment: cfg.commitment,
      chainId: cfg.chainId,
      chain: cfg.chain,
      abi: cfg.abi,
      id: cfg.id,
    };
  }
}

export type HandleEvmConfig = {
  filter: EvmTopicFilter;
  commitment: string;
  chainId: number;
  chain: string;
  abi: string;
  id: string;
};
