import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import {
  decodeNttTransferSent,
  EVMNTTManagerAttributes,
  extractDigestFromNttPayload,
  NTTTransfer,
  SourceChainEvents,
} from "./helpers/ntt";
import { toChainId, ChainId } from "@wormhole-foundation/sdk-base";
import { LogMapperFn, mapLogDataByTopic, mapTxnStatus } from "./helpers/utils";
import {
  AXELAR_SEND_TRANSCEIVER_MESSAGE_ABI,
  WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI,
} from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmSourceChainNttMapper" });

export const evmSourceChainNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const emitterChainId = toChainId(transaction.chainId);
  const transceiverInfo = extractTransceiverInfoAndDigest(transaction.logs, emitterChainId);

  if (!transceiverInfo) {
    logger.warn(
      `[${transaction.chain}] Couldn't map transceiver type: [hash: ${transaction.hash}]`
    );
    return undefined;
  }

  if (!transceiverInfo.digest) {
    logger.warn(`[${transaction.chain}] Couldn't map digest data: [hash: ${transaction.hash}]`);
    return undefined;
  }

  const nttTransferInfo = mapLogDataByTopic(NTT_MANAGER_TOPICS, transaction.logs, emitterChainId);
  const txnStatus = mapTxnStatus(transaction.status);

  if (!nttTransferInfo) {
    logger.warn(`[${transaction.chain}] Couldn't map ntt transfer: [hash: ${transaction.hash}]`);
    return undefined;
  }

  return {
    name: nttTransferInfo.eventName,
    address: transaction.to,
    chainId: emitterChainId,
    blockHeight: BigInt(transaction.blockNumber),
    txHash: transaction.hash.substring(2), // Remove 0x
    blockTime: transaction.timestamp,
    attributes: {
      eventName: nttTransferInfo.eventName,
      from: transaction.from,
      to: transaction.to,
      status: txnStatus,
      blockNumber: transaction.blockNumber,
      timestamp: transaction.timestamp,
      txHash: transaction.hash,
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      gasUsed: transaction.gasUsed,
      effectiveGasPrice: transaction.effectiveGasPrice,
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      protocol: "ntt",
      recipient: nttTransferInfo.recipient,
      amount: nttTransferInfo.amount,
      // We use digest as an unique identifier for the NTT transfer events across source and target chains
      digest: transceiverInfo.digest,
      ...(nttTransferInfo?.fee && {
        fee: nttTransferInfo?.fee,
      }),
      ...(nttTransferInfo?.sourceToken && {
        sourceToken: nttTransferInfo?.sourceToken,
      }),
    },
    tags: {
      recipientChain: nttTransferInfo.recipientChain,
      emitterChain: nttTransferInfo.emitterChain,
      transceiverType: transceiverInfo.transceiverType,
    },
  };
};

export const mapLogDataFromTransferSent: LogMapperFn<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const parsedLog = decodeNttTransferSent(log.data);
  const recipientChainId = toChainId(parsedLog.recipientChain);

  return {
    eventName: "ntt-transfer-sent",
    recipient: parsedLog.recipient,
    amount: BigInt(parsedLog.amount),
    fee: BigInt(parsedLog.fee),
    recipientChain: recipientChainId,
    emitterChain: toChainId(emitterChainId),
    // placeholder, we don't use this returned value
    digest: "digest",
  };
};

type TransceiverLogData = {
  eventName: SourceChainEvents;
  transceiverType: "axelar" | "wormhole";
  recipientChain: ChainId;
  digest: string;
};

export const mapLogDataFromWormholeSendTransceiverMessage: LogMapperFn<TransceiverLogData> = (
  log: EvmTransactionLog,
  emitterChainId: ChainId
): TransceiverLogData | undefined => {
  try {
    const iface = new ethers.utils.Interface(WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI);
    const parsedLog = iface.parseLog(log);

    const nttManagerPayload = parsedLog.args.message.nttManagerPayload;

    const calculatedDigest = extractDigestFromNttPayload(nttManagerPayload, emitterChainId);

    return {
      eventName: "ntt-send-transceiver-message",
      transceiverType: "wormhole",
      recipientChain: toChainId(parsedLog.args.recipientChain),
      digest: calculatedDigest,
    };
  } catch (err) {
    logger.error(`Error parsing wormhole send transceiver message: ${err}`);
  }
};

// SendTransceiverMessage (index_topic_1 uint16 recipientChainId, bytes nttManagerMessage, index_topic_2 bytes32 recipientNttManagerAddress, index_topic_3 bytes32 refundAddress)
export const mapLogDataFromAxelarSendTransceiverMessage: LogMapperFn<TransceiverLogData> = (
  log: EvmTransactionLog,
  emitterChainId: ChainId
): TransceiverLogData | undefined => {
  try {
    const iface = new ethers.utils.Interface(AXELAR_SEND_TRANSCEIVER_MESSAGE_ABI);
    const parsedLog = iface.parseLog(log);

    let nttManagerPayload = parsedLog.args.nttManagerMessage;

    const calculatedDigest = extractDigestFromNttPayload(nttManagerPayload, emitterChainId);

    return {
      eventName: "ntt-send-transceiver-message",
      transceiverType: "axelar",
      recipientChain: toChainId(parsedLog.args.recipientChainId),
      digest: calculatedDigest,
    };
  } catch (err) {
    logger.error(`Error parsing axelar send transceiver message: ${err}`);
  }
};

const NTT_MANAGER_TOPICS: Record<string, LogMapperFn<NTTTransfer>> = {
  "0xe54e51e42099622516fa3b48e9733581c9dbdcb771cafb093f745a0532a35982": mapLogDataFromTransferSent,
};

const AXELAR_SEND_TRANSCEIVER_MESSAGE_TOPIC =
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a";
const WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC =
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738";

const TRANSCEIVER_TOPICS: Record<string, LogMapperFn<TransceiverLogData | undefined>> = {
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a":
    mapLogDataFromAxelarSendTransceiverMessage,
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738":
    mapLogDataFromWormholeSendTransceiverMessage,
};

const extractTransceiverInfoAndDigest = (logs: EvmTransactionLog[], emitterChainId: ChainId) => {
  // check if transceiver is axelar
  const axelarTransceiverLog = logs.find(
    (log) => log.topics[0] === AXELAR_SEND_TRANSCEIVER_MESSAGE_TOPIC
  );
  if (axelarTransceiverLog) {
    const mapper = TRANSCEIVER_TOPICS[AXELAR_SEND_TRANSCEIVER_MESSAGE_TOPIC];
    return mapper(axelarTransceiverLog, emitterChainId);
  }
  const wormholeTransceiverLog = logs.find(
    (log) => log.topics[0] === WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC
  );
  if (wormholeTransceiverLog) {
    const mapper = TRANSCEIVER_TOPICS[WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC];
    return mapper(wormholeTransceiverLog, emitterChainId);
  }
  logger.warn(`Couldn't find transceiver log in transaction logs`);
};
