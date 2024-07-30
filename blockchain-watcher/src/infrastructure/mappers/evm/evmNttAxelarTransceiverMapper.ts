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
import { AXELAR_SEND_TRANSCEIVER_MESSAGE_ABI } from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmNttAxelarTransceiverMapper" });

export const evmNttAxelarTransceiverMapper = (
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

const TRANSCEIVER_TOPICS: Record<string, LogMapperFn<TransceiverLogData | undefined>> = {
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a":
    mapLogDataFromAxelarSendTransceiverMessage,
};
