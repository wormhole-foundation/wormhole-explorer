import { ProcessStandardRelayDelivered } from "./ProcessStandardRelayDelivered";
import { ProcessTransferRedeemed } from "./ProcessTransferRedeemed";
import { ProcessFailedRedeemed } from "./ProcessFailedRedeemed";
import { HandleEvmLogsConfig } from "../HandleEvmLogs";
import { ProcessTransaction } from "./ProcessTransaction";
import { EvmTransaction } from "../../../entities";

export class ProcessTransactionStrategy<T> {
  private mapper: (log: EvmTransaction) => T;
  private transactions: EvmTransaction[];
  private cfg: HandleEvmLogsConfig;

  constructor(
    mapper: (log: EvmTransaction) => T,
    transactions: EvmTransaction[],
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
