import {
  EvmTransactions,
  TransactionFoundEvent,
  StandardRelayDeliveredTransaction,
} from "../../domain/entities";

export const evmStandardRelayDeliveredMapper = (
  transaction: EvmTransactions
): TransactionFoundEvent<StandardRelayDeliveredTransaction> => {
  return {
    name: "standard-relay-delivered",
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
