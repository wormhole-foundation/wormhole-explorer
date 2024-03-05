import { findProtocol } from "../contractsMapper";
import winston from "../../log";
import {
  EvmTransaction,
  EvmTransactionFoundAttributes,
  EvmTransactionLog,
  TransactionFoundEvent,
  TxStatus,
} from "../../../domain/entities";
import { arrayify, hexZeroPad } from "ethers/lib/utils";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

type LogToVaaMapper = (log: EvmTransactionLog) => VaaInformation | undefined;

const mapFromTopics: LogToVaaMapper = (log: EvmTransactionLog) => {
  return {
    emitterChain: Number(log.topics[1]),
    emitterAddress: BigInt(log.topics[2])?.toString(16)?.toUpperCase()?.padStart(64, "0"),
    sequence: Number(log.topics[3]),
  };
};

const mapFromDataBuilder: (dataOffset: number) => LogToVaaMapper = (dataOffset = 0) => {
  return (log: EvmTransactionLog) => {
    const data = Buffer.from(arrayify(log.data, { allowMissingPrefix: true }));

    const offset = dataOffset * 32;
    const emitterChain = data.subarray(offset, offset + 32);
    const emitterAddress = data.subarray(offset + 32, offset + 64);
    const sequence = data.subarray(offset + 64, offset + 96);

    if (emitterChain.length !== 32 || emitterAddress.length !== 32 || sequence.length !== 32) {
      return undefined;
    }

    return {
      emitterChain: emitterChain.readUInt16BE(30),
      emitterAddress: emitterAddress.toString("hex").toUpperCase(),
      sequence: Number(sequence.readBigInt64BE(24)),
    };
  };
};

const RELAYERS: Record<number, string> = {
  10002: "0x7B1bD7a6b4E61c2a123AC6BC2cbfC614437D0470",
  10003: "0x7B1bD7a6b4E61c2a123AC6BC2cbfC614437D0470",
};

const mapFromStandardRelayerDelivery: LogToVaaMapper = (log: EvmTransactionLog) => {
  const emitterChain = Number(log.topics[2]);
  const sourceRelayer = RELAYERS[emitterChain];

  if (!sourceRelayer) return undefined;

  return {
    emitterChain,
    emitterAddress: hexZeroPad(sourceRelayer, 32).substring(2).toUpperCase(),
    sequence: Number(log.topics[3]),
  };
};

const REDEEM_TOPICS: Record<string, LogToVaaMapper> = {
  "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169": mapFromTopics, // Token Bridge
  "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e": mapFromTopics, // CCTP
  "0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0": mapFromDataBuilder(1), // NTT manual
  "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e":
    mapFromStandardRelayerDelivery, // Standard Relayer
};

let logger: winston.Logger;
logger = winston.child({ module: "evmRedeemedTransactionFoundMapper" });

export const evmRedeemedTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EvmTransactionFoundAttributes> | undefined => {
  const first10Characters = transaction.input.slice(0, 10);
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    first10Characters,
    transaction.hash
  );

  const vaaInformation = mappedVaaInformation(transaction.logs);
  const status = mappedStatus(transaction.status);

  const emitterAddress = vaaInformation?.emitterAddress;
  const emitterChain = vaaInformation?.emitterChain;
  const sequence = vaaInformation?.sequence;

  if (protocol && protocol.type && protocol.method) {
    logger.debug(
      `[${transaction.chain}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}][protocol: ${protocol.type}/${protocol.method}]`
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

const mappedVaaInformation = (logs: EvmTransactionLog[]): VaaInformation | undefined => {
  const log = logs.find((log) => {
    return !!REDEEM_TOPICS[log.topics[0]];
  });

  if (!log) return undefined;

  const mapper = REDEEM_TOPICS[log.topics[0]];
  return mapper(log);
};

const mappedStatus = (txStatus: string | undefined): string => {
  switch (txStatus) {
    case TX_STATUS_CONFIRMED:
      return TxStatus.Confirmed;
    case TX_STATUS_FAILED:
      return TxStatus.Failed;
    default:
      return TxStatus.Unkonwn;
  }
};

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};
