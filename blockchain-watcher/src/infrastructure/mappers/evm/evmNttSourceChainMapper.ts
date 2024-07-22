import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import {
  decodeNttTransferSent,
  EVMNTTManagerAttributes,
  getNttManagerMessageDigest,
  NativeTokenTransfer,
  NttManagerMessage,
  NTTTransfer,
  WormholeTransceiverMessage,
} from "./helpers/ntt";
import { toChainId, chainIdToChain, ChainId } from "@wormhole-foundation/sdk-base";
import { UniversalAddress } from "@wormhole-foundation/sdk-definitions";
import { LogToNTTTransfer, mapLogDataByTopic, mappedTxnStatus } from "./helpers/utils";

let logger: winston.Logger = winston.child({ module: "evmSourceChainNttMapper" });

export const evmSourceChainNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const transceiverInfo = mapLogDataByTopic(TRANSCEIVER_TOPICS, transaction.logs);

  if (!transceiverInfo) {
    logger.warn(
      `[${transaction.chain}] Couldn't map transceiver type: [hash: ${transaction.hash}]`
    );
    return undefined;
  }

  const digest = mapLogDataByTopic(
    LOG_MESSAGE_PUBLISHED_TOPIC,
    transaction.logs,
    transaction.chainId
  );

  if (!digest) {
    logger.warn(`[${transaction.chain}] Couldn't map digest data: [hash: ${transaction.hash}]`);
    return undefined;
  }

  const nttTransferInfo = mapLogDataByTopic(TOPICS, transaction.logs, transaction.chainId);
  const txnStatus = mappedTxnStatus(transaction.status);

  if (!nttTransferInfo) {
    logger.warn(`[${transaction.chain}] Couldn't map ntt transfer: [hash: ${transaction.hash}]`);
    return undefined;
  }

  return {
    name: nttTransferInfo.eventName,
    address: transaction.to,
    chainId: transaction.chainId,
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
      digest,
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

export const mapLogDataFromTransferSent: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const parsedLog = decodeNttTransferSent(log.data);
  const recipientChainId = toChainId(parsedLog.recipientChain);

  return {
    eventName: "transfer-sent",
    recipient: new UniversalAddress(parsedLog.recipient).toNative(chainIdToChain(recipientChainId)),
    amount: BigInt(parsedLog.amount),
    fee: BigInt(parsedLog.fee),
    recipientChain: recipientChainId,
    emitterChain: toChainId(emitterChainId),
    // placeholder, we don't use this returned value
    digest: "digest",
  };
};

type TransceiverLogData = {
  eventName: string;
  transceiverType: "axelar" | "wormhole";
  recipientChain: ChainId;
};

export const mapLogDataFromWormholeSendTransceiverMessage: LogToNTTTransfer<TransceiverLogData> = (
  log: EvmTransactionLog
): TransceiverLogData => {
  const abi = "event SendTransceiverMessage(uint16 recipientChain, tuple message)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "send-transceiver-message",
    transceiverType: "wormhole",
    recipientChain: toChainId(parsedLog.args.recipientChain),
  };
};

export const mapLogDataFromAxelarSendTransceiverMessage: LogToNTTTransfer<TransceiverLogData> = (
  log: EvmTransactionLog
): TransceiverLogData => {
  const abi =
    "event SendTransceiverMessage(index_topic_1 uint16 recipientChainId, bytes nttManagerMessage, index_topic_2 bytes32 recipientNttManagerAddress, index_topic_3 bytes32 refundAddress)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "send-transceiver-message",
    transceiverType: "axelar",
    recipientChain: toChainId(parsedLog.args.recipientChainId),
  };
};

export const mapLogDataFromMessagePublished: LogToNTTTransfer<string> = (
  log: EvmTransactionLog,
  emitterChainId: number
): string => {
  const abi =
    "event LogMessagePublished (index_topic_1 address sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  let payload = parsedLog.args.payload;

  // Strip off leading 0x, if present
  if (payload.startsWith("0x")) {
    payload = payload.slice(2);
  }

  const payloadBuffer = Buffer.from(payload, "hex");

  const transceiverMessage = WormholeTransceiverMessage.deserialize(payloadBuffer, (a) =>
    NttManagerMessage.deserialize(a, NativeTokenTransfer.deserialize)
  );

  const calculatedDigest = getNttManagerMessageDigest(
    emitterChainId,
    transceiverMessage.ntt_managerPayload
  );

  return calculatedDigest;
};

const TOPICS: Record<string, LogToNTTTransfer<NTTTransfer>> = {
  "0xe54e51e42099622516fa3b48e9733581c9dbdcb771cafb093f745a0532a35982": mapLogDataFromTransferSent,
};

const TRANSCEIVER_TOPICS: Record<string, LogToNTTTransfer<TransceiverLogData>> = {
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738":
    mapLogDataFromWormholeSendTransceiverMessage,
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a":
    mapLogDataFromAxelarSendTransceiverMessage,
};

const LOG_MESSAGE_PUBLISHED_TOPIC: Record<string, LogToNTTTransfer<string>> = {
  "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2":
    mapLogDataFromMessagePublished,
};
