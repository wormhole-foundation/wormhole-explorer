import { EvmTransaction, EvmTransactionFound } from "../../entities";
import { HandleEvmLogsConfig } from "./HandleEvmLogs";

/**
 * Handling means mapping and forward to a given target.
 * As of today, we have mapped this event evmFailedRedeemed, evmStandardRelayDelivered and evmTransferRedeemed.
 */
export class HandleEvmTransactions<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (log: EvmTransaction) => T;
  target: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (log: EvmTransaction) => T,
    target: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = this.normalizeCfg(cfg);
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(transactions: EvmTransaction[]): Promise<T[]> {
    const mappedItems = transactions.map((transaction) => {
      return this.mapper(transaction);
    }) as EvmTransactionFound[];

    const filterItems = mappedItems.filter(
      (transaction) => transaction.methodsByAddress || transaction.name
    ) as T[];

    await this.target(filterItems);

    // TODO: return a result specifying failures if any
    return filterItems;
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
