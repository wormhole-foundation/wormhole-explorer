import { arrayify, hexZeroPad } from "ethers/lib/utils";
import { STANDARD_RELAYERS } from "../../constants";
import { configuration } from "../../config";
import { findProtocol } from "../contractsMapper";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "../../log";
import {
  EvmTransactionFoundAttributes,
  TransactionFoundEvent,
  EvmTransactionLog,
  EvmTransaction,
  TxStatus,
} from "../../../domain/entities";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

let logger: winston.Logger;
logger = winston.child({ module: "evmRedeemedTransactionFoundMapper" });

export const evmRedeemedTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EvmTransactionFoundAttributes> | undefined => {
  const vaaInformation = mappedVaaInformation(transaction.logs, transaction.input);
  const status = mappedStatus(transaction.status);

  // Validate correct vaa information
  if (!vaaInformation || vaaInformation.emitterChain === 0) {
    logger.warn(
      `[${transaction.chain}] Cannot mapper vaa information: [hash: ${
        transaction.hash
      }][VAA: ${JSON.stringify(vaaInformation)}]`
    );
    return undefined;
  }

  const first10Characters = transaction.input.slice(0, 10);
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    first10Characters,
    transaction.hash
  );
  const { type: protocolType, method: protocolMethod } = protocol;

  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const sequence = vaaInformation.sequence;

  logger.debug(
    `[${transaction.chain}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}][protocol: ${protocolType}/${protocolMethod}]`
  );

  return {
    name: "transfer-redeemed",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash.substring(2), // Remove 0x
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      from: transaction.from,
      to: transaction.to,
      status: status,
      blockNumber: transaction.blockNumber,
      input: transaction.input,
      methodsByAddress: protocolMethod,
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
      protocol: protocolType,
    },
  };
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

/**
 * Mapped vaa information from logs.data or input using the topics to map the correct mapper
 */
const mappedVaaInformation = (
  logs: EvmTransactionLog[],
  input: string
): VaaInformation | undefined => {
  const filterLogs = logs.filter((log) => {
    return REDEEM_TOPICS[log.topics[0]];
  });

  if (!filterLogs) return undefined;

  for (const log of filterLogs) {
    const mapper = REDEEM_TOPICS[log.topics[0]];
    const vaaInformation = mapper(log, input);

    if (
      vaaInformation &&
      vaaInformation.emitterChain != 0 &&
      vaaInformation.emitterAddress &&
      vaaInformation.sequence
    ) {
      return vaaInformation;
    }
  }
};

const mapVaaFromTopics: LogToVaaMapper = (log: EvmTransactionLog) => {
  return {
    emitterChain: Number(log.topics[1]),
    emitterAddress: BigInt(log.topics[2])?.toString(16)?.toUpperCase()?.padStart(64, "0"),
    sequence: Number(log.topics[3]),
  };
};

const mapVaaFromDataBuilder: (dataOffset: number) => LogToVaaMapper = (dataOffset = 0) => {
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

const mapVaaFromStandardRelayerDelivery: LogToVaaMapper = (log: EvmTransactionLog) => {
  const emitterChain = Number(log.topics[2]);
  const sourceRelayer = STANDARD_RELAYERS[configuration.environment][emitterChain];

  if (!sourceRelayer) return undefined;

  return {
    emitterChain,
    emitterAddress: hexZeroPad(sourceRelayer, 32).substring(2).toUpperCase(),
    sequence: Number(log.topics[3]),
  };
};

const mapVaaFromInput: LogToVaaMapper = (_, input: string) => {
  const vaaBuffer = Buffer.from(input.substring(138), "hex");
  const vaa = parseVaa(vaaBuffer);

  return {
    emitterAddress: vaa.emitterAddress.toString("hex"),
    emitterChain: vaa.emitterChain,
    sequence: Number(vaa.sequence),
  };
};

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};

type LogToVaaMapper = (log: EvmTransactionLog, input: string) => VaaInformation | undefined;

const REDEEM_TOPICS: Record<string, LogToVaaMapper> = {
  "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169": mapVaaFromTopics, // Token Bridge
  "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": mapVaaFromInput, // Token Bridge sepolia
  "0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0": mapVaaFromDataBuilder(1), // NTT manual
  "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e": mapVaaFromTopics, // CCTP
  "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e":
    mapVaaFromStandardRelayerDelivery, // Standard Relayer
};
