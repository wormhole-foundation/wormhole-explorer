import { ProcessTransaction } from "./ProcessTransaction";
import { EvmTransaction } from "../../../entities";

const MAPPER_NAME = "evmFailedRedeemedMapper";
const STATUS_FAIL = "0x0";

export class ProcessFailedRedeemed<T> implements ProcessTransaction<T> {
  private mapper: (log: EvmTransaction) => T;
  private transactions: EvmTransaction[];

  constructor(mapper: (log: EvmTransaction) => T, transactions: EvmTransaction[]) {
    this.transactions = transactions;
    this.mapper = mapper;
  }

  apply(): boolean {
    return this.mapper.name == MAPPER_NAME;
  }

  execute(): T[] {
    return this.transactions
      .filter((transaction) => transaction.status === STATUS_FAIL && transaction.methodsByAddress)
      .map((transaction) => {
        return this.mapper(transaction);
      });
  }
}
