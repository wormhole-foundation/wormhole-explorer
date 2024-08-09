import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import {
  EVMNTTManagerAttributes,
  extractDigestFromNttPayload,
  NTTTransfer,
  parseNttPayload,
} from "./helpers/ntt";
import { toChainId, ChainId } from "@wormhole-foundation/sdk-base";
import {
  isTopicPresentInLogs,
  LogMapperFn,
  mapLogDataByTopic,
  mapTxnStatus,
} from "./helpers/utils";
import { WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI } from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmNttWormholeTransceiverMapper" });

export const evmNttWormholeTransceiverMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const emitterChainId = toChainId(transaction.chainId);

  // Process further only if transaction logs contain TRANSCEIVER_TOPICS
  if (!isTopicPresentInLogs(TRANSCEIVER_TOPICS, transaction.logs)) {
    return undefined;
  }

  const transceiverInfo = mapLogDataByTopic(TRANSCEIVER_TOPICS, transaction.logs, emitterChainId);
  const txnStatus = mapTxnStatus(transaction.status);

  if (!transceiverInfo) {
    logger.warn(`[${transaction.chain}] Couldn't map ntt transfer: [hash: ${transaction.hash}]`);
    return undefined;
  }

  return {
    name: transceiverInfo.eventName,
    address: transaction.to,
    chainId: emitterChainId,
    blockHeight: BigInt(transaction.blockNumber),
    txHash: transaction.hash.substring(2), // Remove 0x
    blockTime: transaction.timestamp,
    attributes: {
      eventName: transceiverInfo.eventName,
      from: transaction.from,
      to: transaction.to,
      status: txnStatus,
      blockNumber: transaction.blockNumber,
      timestamp: transaction.timestamp,
      txHash: transaction.hash,
      gas: BigInt(transaction.gas),
      gasPrice: BigInt(transaction.gasPrice),
      gasUsed: BigInt(transaction.gasUsed),
      effectiveGasPrice: BigInt(transaction.effectiveGasPrice),
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      protocol: "NTT",
      // We use digest as an unique identifier for the NTT transfer events across source and target chains
      digest: transceiverInfo.digest,
      amount: transceiverInfo?.amount,
    },
    tags: {
      recipientChain: transceiverInfo.recipientChain,
      emitterChain: emitterChainId,
      transceiverType: transceiverInfo.transceiverType,
      sourceToken: transceiverInfo?.sourceToken,
    },
  };
};

export const mapLogDataFromWormholeSendTransceiverMessage: LogMapperFn<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: ChainId
): NTTTransfer | undefined => {
  try {
    const iface = new ethers.utils.Interface(WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI);
    const parsedLog = iface.parseLog(log);

    const nttManagerPayload = parsedLog.args.message.nttManagerPayload;
    const calculatedDigest = extractDigestFromNttPayload(nttManagerPayload, emitterChainId);
    const parsedNttPayload = parseNttPayload(nttManagerPayload);

    return {
      eventName: "ntt-send-transceiver-message",
      transceiverType: "wormhole",
      recipientChain: toChainId(parsedLog.args.recipientChain),
      digest: calculatedDigest,
      sourceToken: Buffer.from(parsedNttPayload?.payload?.sourceToken).toString("hex"),
      amount: parsedNttPayload?.payload?.trimmedAmount?.amount,
    };
  } catch (err) {
    logger.error(`Error parsing wormhole send transceiver message: ${err}`);
  }
};

const TRANSCEIVER_TOPICS: Record<string, LogMapperFn<NTTTransfer | undefined>> = {
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738":
    mapLogDataFromWormholeSendTransceiverMessage,
};
