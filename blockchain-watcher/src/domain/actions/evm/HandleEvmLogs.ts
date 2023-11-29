import { ethers } from "ethers";
import { EvmLog, EvmTopicFilter } from "../../entities";

/**
 * Handling means mapping and forward to a given target.
 * As of today, only one type of event can be handled per each instance.
 */
export class HandleEvmLogs<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (log: EvmLog, parsedArgs: ReadonlyArray<any>, cfg: HandleEvmLogsConfig) => T;
  target: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmLog, args: ReadonlyArray<any>, cfg: HandleEvmLogsConfig) => T,
    target: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = this.normalizeCfg(cfg);
    this.mapper = mapper;
    this.target = target;
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
        return this.mapper(log, parsedLog.args, this.cfg);
      });

    await this.target(mappedItems);
    // TODO: return a result specifying failures if any
    return mappedItems;
  }

  private normalizeCfg(cfg: HandleEvmLogsConfig): HandleEvmLogsConfig {
    return {
      filter: {
        addresses: cfg.filter.addresses.map((addr) => addr.toLowerCase()),
        topics: cfg.filter.topics.map((topic) => topic.toLowerCase()),
      },
      abi: cfg.abi,
      chainId: cfg.chainId,
    };
  }
}

export type HandleEvmLogsConfig = {
  filter: EvmTopicFilter;
  abi: string;
  chainId: number;
};
