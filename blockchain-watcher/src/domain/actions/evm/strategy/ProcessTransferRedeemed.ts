import { HandleEvmLogsConfig } from "../HandleEvmLogs";
import { ProcessTransaction } from "./ProcessTransaction";
import { EvmTransactions } from "../../../entities";

const STATUS_SUCCESS = "0x1";
const MAPPER_NAME = "evmTransferRedeemedMapper";

export class ProcessTransferRedeemed<T> implements ProcessTransaction<T> {
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

  apply(): boolean {
    const mapperName = this.mapper.name;
    return mapperName == MAPPER_NAME;
  }

  execute(): T[] {
    return this.transactions
      .filter(
        (transaction) =>
          this.cfg.filter.addresses.includes(transaction.to.toLowerCase()) &&
          transaction.status === STATUS_SUCCESS
      )
      .map((transaction) => {
        return this.mapper(transaction);
      });
  }
}
