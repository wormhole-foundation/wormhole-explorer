import {
  EvmTransactions,
  FailedRedeemedTransaction,
  TransactionFoundEvent,
} from "../../domain/entities";

export const evmFailedRedeemedMapper = (
  transaction: EvmTransactions
): TransactionFoundEvent<FailedRedeemedTransaction> => {
  return {
    name: "failed-redeemed",
    address: transaction.to,
    chainId: Number(transaction.chainId),
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    attributes: {
      from: transaction.from,
      to: transaction.to,
      status: transaction.status,
      blockNumber: transaction.blockNumber,
      input: transaction.input,
      methodsByAddress: transaction.methodsByAddress,
    },
  };
};