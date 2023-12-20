import { ProcessTransaction } from "./ProcessTransactionStrategy";
import { EvmTransactions } from "../../../entities";

const MAPPER_NAME = "evmFailedRedeemedMapper";
const STATUS_FAIL = "0x0";

export class ProcessFailedRedeemed<T> implements ProcessTransaction<T> {
  private mapper: (log: EvmTransactions) => T;
  private transactions: EvmTransactions[];

  constructor(mapper: (log: EvmTransactions) => T, transactions: EvmTransactions[]) {
    this.transactions = transactions;
    this.mapper = mapper;
  }

  apply(): boolean {
    return this.mapper.name == MAPPER_NAME;
  }

  execute(): T[] {
    return this.transactions
      .filter((transaction) => transaction.status === STATUS_FAIL)
      .map((transaction) => {
        return this.mapper(transaction);
      });
  }
}
