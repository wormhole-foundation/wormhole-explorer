import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import {
  EVMNTTManagerAttributes,
  extractDigestFromNttPayload,
  TransceiverLogData,
} from "./helpers/ntt";
import { toChainId, ChainId } from "@wormhole-foundation/sdk-base";
import { LogMapperFn, mapLogDataByTopic, mapTxnStatus } from "./helpers/utils";
import { WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI } from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmNttWormholeTransceiverMapper" });

export const evmNttWormholeTransceiverMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const emitterChainId = toChainId(transaction.chainId);

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
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      gasUsed: transaction.gasUsed,
      effectiveGasPrice: transaction.effectiveGasPrice,
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      protocol: "NTT",
      // We use digest as an unique identifier for the NTT transfer events across source and target chains
      digest: transceiverInfo.digest,
    },
    tags: {
      recipientChain: transceiverInfo.recipientChain,
      emitterChain: emitterChainId,
      transceiverType: transceiverInfo.transceiverType,
    },
  };
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

const TRANSCEIVER_TOPICS: Record<string, LogMapperFn<TransceiverLogData | undefined>> = {
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738":
    mapLogDataFromWormholeSendTransceiverMessage,
};
