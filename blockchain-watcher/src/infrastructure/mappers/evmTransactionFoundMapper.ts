import { methodNameByAddressMapper } from "../../domain/actions/evm/mappers/methodNameByAddressMapper";
import { EvmTransaction, TransactionFound, TransactionFoundEvent } from "../../domain/entities";

export const evmTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<TransactionFound> => {
  const protocol = methodNameByAddressMapper(
    transaction.chain,
    transaction.environment,
    transaction
  );

  return {
    name: "evm-transaction-found",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    attributes: {
      name: protocol?.name,
      from: transaction.from,
      to: transaction.to,
      status: transaction.status,
      blockNumber: transaction.blockNumber,
      input: transaction.input,
      methodsByAddress: protocol?.method,
      timestamp: transaction.timestamp,
      blockHash: transaction.blockHash,
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      maxFeePerGas: transaction.maxFeePerGas,
      maxPriorityFeePerGas: transaction.maxPriorityFeePerGas,
      nonce: transaction.nonce,
      r: transaction.r,
      s: transaction.s,
      transactionIndex: transaction.transactionIndex,
      type: transaction.type,
      v: transaction.v,
      value: transaction.value,
    },
  };
};
