import { EvmTransaction, TransactionFound, TransactionFoundEvent } from "../../../domain/entities";
import { methodNameByAddressMapper } from "./methodNameByAddressMapper";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

const TOKEN_BRIDGE_TOPIC = "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169";
const CCTP_TOPIC = "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e";

export const evmRedeemedTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<TransactionFound> => {
  const protocol = methodNameByAddressMapper(
    transaction.chain,
    transaction.environment,
    transaction
  );

  const vaaInformation = mappedVAAinformation(transaction);
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
      sequence: vaaInformation?.sequence,
      emitterAddress: vaaInformation?.emitterAddress,
      emitterChain: vaaInformation?.emitterChain,
    },
  };
};

const mappedVAAinformation = (transaction: EvmTransaction): vaaInformation | undefined => {
  const vaaInformation: vaaInformation = {};
  const logs = transaction.logs;

  logs
    .filter((log) => {
      return log.topics?.includes(CCTP_TOPIC) || log.topics?.includes(TOKEN_BRIDGE_TOPIC);
    })
    .map((log) => {
      (vaaInformation.emitterChain = Number(log.topics[1])),
        (vaaInformation.emitterAddress = BigInt(log.topics[2])
          .toString(16)
          .toUpperCase()
          .padStart(64, "0")),
        (vaaInformation.sequence = Number(log.topics[3]));
    });

  return vaaInformation;
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

type vaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};

export enum status {
  TxStatusConfirmed = "completed",
  TxStatusUnkonwn = "unknown",
  TxStatusFailed = "failed",
}
