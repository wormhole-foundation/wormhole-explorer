import { ProcessStandardRelayDelivered } from "./ProcessStandardRelayDelivered";
import { ProcessTransferRedeemed } from "./ProcessTransferRedeemed";
import { ProcessFailedRedeemed } from "./ProcessFailedRedeemed";
import { HandleEvmLogsConfig } from "../HandleEvmLogs";
import { ProcessTransaction } from "./ProcessTransaction";
import { EvmTransactions } from "../../../entities";

export class ProcessTransactionStrategy<T> {
  private mapper: (log: EvmTransactions) => T;
  private transactions: EvmTransactions[];
  private cfg: HandleEvmLogsConfig;

  constructor(
    mapper: (log: EvmTransactions) => T,
    transactions: EvmTransactions[],
    cfg: HandleEvmLogsConfig
  ) {
    this.transactions = transactions;
    this.mapper = mapper;
    this.cfg = cfg;
  }

  execute(): T[] {
    let mappedItems: T[] = [];

    const processors: ProcessTransaction<T>[] = [
      new ProcessFailedRedeemed(this.mapper, this.transactions),
      new ProcessStandardRelayDelivered(this.mapper, this.transactions, this.cfg),
      new ProcessTransferRedeemed(this.mapper, this.transactions, this.cfg),
    ];

    processors.forEach((process) => {
      if (process.apply()) {
        mappedItems = process.execute();
        return;
      }
    });

    return mappedItems;
  }
}
