import {
  EvmTransactions,
  TransactionFoundEvent,
  TransferRedeemedTransaction,
} from "../../domain/entities";

export const evmTransferRedeemedMapper = (
  transaction: EvmTransactions
): TransactionFoundEvent<TransferRedeemedTransaction> => {
  return {
    name: "transfer-redeemed",
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
