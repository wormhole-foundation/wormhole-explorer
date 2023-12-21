import { HandleEvmLogsConfig } from "../HandleEvmLogs";
import { ProcessTransaction } from "./ProcessTransaction";
import { EvmTransaction } from "../../../entities";

const STATUS_SUCCESS = "0x1";
const MAPPER_NAME = "evmStandardRelayDeliveredMapper";

export class ProcessStandardRelayDelivered<T> implements ProcessTransaction<T> {
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

  apply(): boolean {
    return this.mapper.name == MAPPER_NAME;
  }

  execute(): T[] {
    return this.transactions
      .filter(
        (transaction) =>
          this.cfg.filter.addresses.includes(transaction.to.toLowerCase()) &&
          transaction.status === STATUS_SUCCESS &&
          transaction.methodsByAddress
      )
      .map((transaction) => {
        return this.mapper(transaction);
      });
  }
}
