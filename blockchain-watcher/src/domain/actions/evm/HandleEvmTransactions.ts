import { EvmTopicFilter, EvmTransactions } from "../../entities";
import { ethers } from "ethers";

/**
 * Handling means mapping and forward to a given target.
 * As of today, only one type of event can be handled per each instance.
 */
export class HandleEvmTransactions<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (log: EvmTransactions, parsedArgs: ReadonlyArray<any>) => T;
  target: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmTransactions, args: ReadonlyArray<any>) => T,
    target: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = this.normalizeCfg(cfg);
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(transactions: EvmTransactions[]): Promise<T[]> {
    const mappedItems = transactions.map((transaction) => {
      const iface = new ethers.utils.Interface([this.cfg.abi]);
      const parsedLog = iface.parseLog(transaction);
      return this.mapper(transaction, parsedLog.args);
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
    };
  }
}

export type HandleEvmLogsConfig = {
  filter: EvmTopicFilter;
  abi: string;
};
