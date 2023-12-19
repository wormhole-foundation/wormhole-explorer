import { ProcessTransactionStrategy } from "./strategy/ProcessTransactionStrategy";
import { HandleEvmLogsConfig } from "./HandleEvmLogs";
import { EvmTransactions } from "../../entities";

/**
 * Handling means mapping and forward to a given target.
 * As of today, only one type of event can be handled per each instance.
 */
export class HandleEvmTransactions<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (log: EvmTransactions) => T;
  target: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmTransactions) => T,
    target: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = this.normalizeCfg(cfg);
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(transactions: EvmTransactions[]): Promise<T[]> {
    const mappedItems = new ProcessTransactionStrategy(
      this.mapper,
      transactions,
      this.cfg
    ).execute();
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