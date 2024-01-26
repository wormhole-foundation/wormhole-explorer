import { EvmTransaction, EvmTransactionFoundAttributes, TransactionFound, TransactionFoundEvent } from "../../../domain/entities";
import { Protocol, contractsMapperConfig } from "../contractsMapper";
import winston from "../../log";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

const TOKEN_BRIDGE_TOPIC = "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169";
const CCTP_TOPIC = "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e";

let logger: winston.Logger;
logger = winston.child({ module: "evmRedeemedTransactionFoundMapper" });

export const evmRedeemedTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EvmTransactionFoundAttributes> | undefined => {
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    transaction.input,
    transaction.hash
  );

  const vaaInformation = mappedVaaInformation(transaction.logs);
  const status = mappedStatus(transaction.status);

  const emitterAddress = vaaInformation?.emitterAddress;
  const emitterChain = vaaInformation?.emitterChain;
  const sequence = vaaInformation?.sequence;

  if (protocol && protocol.type && protocol.method) {
    logger.info(
      `[${transaction.chain}][evmRedeemedTransactionFoundMapper] Transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
    );

    return {
      name: "transfer-redeemed",
      address: transaction.to,
      chainId: transaction.chainId,
      txHash: transaction.hash,
      blockHeight: BigInt(transaction.blockNumber),
      blockTime: transaction.timestamp,
      attributes: {
        from: transaction.from,
        to: transaction.to,
        status: status,
        blockNumber: transaction.blockNumber,
        input: transaction.input,
        methodsByAddress: protocol.method,
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
        protocol: protocol.type,
      },
    };
  }
};

const mappedVaaInformation = (
  logs: { address: string; topics: string[] }[]
): VaaInformation | undefined => {
  const log = logs.find((log) => {
    if (log.topics.includes(CCTP_TOPIC) || log.topics.includes(TOKEN_BRIDGE_TOPIC)) return log;
  });

  const vaaInformation = log
    ? {
        emitterChain: Number(log.topics[1]),
        emitterAddress: BigInt(log.topics[2])?.toString(16)?.toUpperCase()?.padStart(64, "0"),
        sequence: Number(log.topics[3]),
      }
    : undefined;

  return vaaInformation;
};

const mappedStatus = (txStatus: string | undefined): string => {
  switch (txStatus) {
    case TX_STATUS_CONFIRMED:
      return status.TxStatusConfirmed;
    case TX_STATUS_FAILED:
      return status.TxStatusFailed;
    default:
      return status.TxStatusUnkonwn;
  }
};

const findProtocol = (
  chain: string,
  address: string,
  input: string,
  hash: string
): Protocol | undefined => {
  const first10Characters = input.slice(0, 10);

  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === chain) {
      const foundProtocol = contract.protocols.find((protocol) =>
        protocol.addresses.includes(address)
      );
      const foundMethod = foundProtocol?.methods.find(
        (method) => method.methodId === first10Characters
      );

      if (foundMethod && foundProtocol) {
        return {
          method: foundMethod.method,
          type: foundProtocol.type,
        };
      }
    }
  }

  logger.warn(
    `[${chain}] Protocol not found, [tx hash: ${hash}][address: ${address}][input: ${input}]`
  );
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
