import {
  EvmTransaction,
  FailedRedeemedTransaction,
  TransactionFoundEvent,
} from "../../domain/entities";

export const evmFailedRedeemedMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<FailedRedeemedTransaction> => {
  return {
    name: "redeemed-failed",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    attributes: {
      from: transaction.from,
      to: transaction.to,
      status: transaction.status,
      blockNumber: transaction.blockNumber,
      input: transaction.input,
      methodsByAddress: transaction.methodsByAddress,
      timestamp: transaction.timestamp,
    },
  };
};
