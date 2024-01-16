import { methodNameByAddressMapper } from "./methodNameByAddressMapper";
import {
  EvmTransaction,
  EvmTransactionFound,
  TransactionFoundEvent,
} from "../../../domain/entities";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

export const evmTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EvmTransactionFound> => {
  const protocol = methodNameByAddressMapper(
    transaction.chain,
    transaction.environment,
    transaction
  );

  const status = mappedStatus(transaction);

  return {
    name: "evm-transaction-found",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      name: protocol?.name,
      from: transaction.from,
      to: transaction.to,
      status: status,
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
      sequence: transaction.sequence,
      emitterAddress: transaction.emitterAddress,
      emitterChain: transaction.emitterChain,
    },
  };
};

const mappedStatus = (transaction: EvmTransaction): string => {
  switch (transaction.status) {
    case TX_STATUS_CONFIRMED:
      return status.TxStatusConfirmed;
    case TX_STATUS_FAILED:
      return status.TxStatusFailed;
    default:
      return status.TxStatusUnkonwn;
  }
};

export enum status {
  TxStatusConfirmed = "completed",
  TxStatusUnkonwn = "unknown",
  TxStatusFailed = "failed",
}
