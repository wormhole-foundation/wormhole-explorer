import { EvmTransaction, TransactionFound, TransactionFoundEvent } from "../../../domain/entities";
import { methodNameByAddressMapper } from "./methodNameByAddressMapper";
import winston from "../../log";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

const TOKEN_BRIDGE_TOPIC = "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169";
const CCTP_TOPIC = "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e";

let logger: winston.Logger;
logger = winston.child({ module: "evmRedeemedTransactionFoundMapper" });

export const evmRedeemedTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<TransactionFound> => {
  const protocol = methodNameByAddressMapper(
    transaction.chain,
    transaction.environment,
    transaction
  );

  const vaaInformation = mappedVaaInformation(transaction);
  const status = mappedStatus(transaction);

  const emitterAddress = vaaInformation?.emitterAddress;
  const emitterChain = vaaInformation?.emitterChain;
  const sequence = vaaInformation?.sequence;

  logger.info(
    `[${transaction.chain}][evmRedeemedTransactionFoundMapper] Transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

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
      sequence: sequence,
      emitterAddress: emitterAddress,
      emitterChain: emitterChain,
    },
  };
};

const mappedVaaInformation = (transaction: EvmTransaction): VaaInformation | undefined => {
  const logs = transaction.logs;

  const log = logs.find((log) => {
    if (log.topics.includes(CCTP_TOPIC) || log.topics.includes(TOKEN_BRIDGE_TOPIC)) return log;
  });

  const vaaInformation = {
    emitterChain: Number(log?.topics[1]),
    emitterAddress: log?.topics[2]
      ? BigInt(log.topics[2])?.toString(16)?.toUpperCase()?.padStart(64, "0")
      : undefined,
    sequence: Number(log?.topics[3]),
  };

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

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};

export enum status {
  TxStatusConfirmed = "completed",
  TxStatusUnkonwn = "unknown",
  TxStatusFailed = "failed",
}
