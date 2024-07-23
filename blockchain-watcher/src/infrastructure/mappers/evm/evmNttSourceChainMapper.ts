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
} from "./helpers/ntt";
import { toChainId, ChainId } from "@wormhole-foundation/sdk-base";
import { LogToNTTTransfer, mapLogDataByTopic, mappedTxnStatus } from "./helpers/utils";

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

  const nttTransferInfo = mapLogDataByTopic(MAIN_TOPICS, transaction.logs, emitterChainId);
  const txnStatus = mappedTxnStatus(transaction.status);

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

export const mapLogDataFromTransferSent: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const parsedLog = decodeNttTransferSent(log.data);
  const recipientChainId = toChainId(parsedLog.recipientChain);

  return {
    eventName: "transfer-sent",
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
  eventName: string;
  transceiverType: "axelar" | "wormhole";
  recipientChain: ChainId;
  digest: string;
};

export const mapLogDataFromWormholeSendTransceiverMessage: LogToNTTTransfer<TransceiverLogData> = (
  log: EvmTransactionLog,
  emitterChainId: number
): TransceiverLogData | undefined => {
  try {
    const abi = [
      {
        anonymous: false,
        inputs: [
          { indexed: false, internalType: "uint16", name: "recipientChain", type: "uint16" },
          {
            components: [
              { internalType: "bytes32", name: "sourceNttManagerAddress", type: "bytes32" },
              { internalType: "bytes32", name: "recipientNttManagerAddress", type: "bytes32" },
              { internalType: "bytes", name: "nttManagerPayload", type: "bytes" },
              { internalType: "bytes", name: "transceiverPayload", type: "bytes" },
            ],
            indexed: false,
            internalType: "struct TransceiverStructs.TransceiverMessage",
            name: "message",
            type: "tuple",
          },
        ],
        name: "SendTransceiverMessage",
        type: "event",
      },
    ];
    const iface = new ethers.utils.Interface(abi);
    const parsedLog = iface.parseLog(log);

    let nttManagerPayload = parsedLog.args.message.nttManagerPayload;

    // Strip off leading 0x, if present
    if (nttManagerPayload.startsWith("0x")) {
      nttManagerPayload = nttManagerPayload.slice(2);
    }

    const payloadBuffer = Buffer.from(nttManagerPayload, "hex");

    const nttPayload = NttManagerMessage.deserialize(
      payloadBuffer,
      NativeTokenTransfer.deserialize
    );

    const calculatedDigest = getNttManagerMessageDigest(emitterChainId, nttPayload);

    return {
      eventName: "send-transceiver-message",
      transceiverType: "wormhole",
      recipientChain: toChainId(parsedLog.args.recipientChain),
      digest: calculatedDigest,
    };
  } catch (err) {
    logger.error(`Error parsing wormhole send transceiver message: ${err}`);
  }
};

// SendTransceiverMessage (index_topic_1 uint16 recipientChainId, bytes nttManagerMessage, index_topic_2 bytes32 recipientNttManagerAddress, index_topic_3 bytes32 refundAddress)
export const mapLogDataFromAxelarSendTransceiverMessage: LogToNTTTransfer<TransceiverLogData> = (
  log: EvmTransactionLog,
  emitterChainId: ChainId
): TransceiverLogData | undefined => {
  try {
    // abi ref: https://sepolia.etherscan.io/address/0xcc6e5c994de73e8a115263b1b512e29b2026df55#code
    const abi = [
      {
        anonymous: false,
        inputs: [
          { indexed: true, internalType: "uint16", name: "recipientChainId", type: "uint16" },
          { indexed: false, internalType: "bytes", name: "nttManagerMessage", type: "bytes" },
          {
            indexed: true,
            internalType: "bytes32",
            name: "recipientNttManagerAddress",
            type: "bytes32",
          },
          { indexed: true, internalType: "bytes32", name: "refundAddress", type: "bytes32" },
        ],
        name: "SendTransceiverMessage",
        type: "event",
      },
    ];

    const iface = new ethers.utils.Interface(abi);
    const parsedLog = iface.parseLog(log);

    let nttManagerPayload = parsedLog.args.nttManagerMessage;

    // Strip off leading 0x, if present
    if (nttManagerPayload.startsWith("0x")) {
      nttManagerPayload = nttManagerPayload.slice(2);
    }

    const payloadBuffer = Buffer.from(nttManagerPayload, "hex");

    const nttPayload = NttManagerMessage.deserialize(
      payloadBuffer,
      NativeTokenTransfer.deserialize
    );

    const calculatedDigest = getNttManagerMessageDigest(emitterChainId, nttPayload);

    return {
      eventName: "send-transceiver-message",
      transceiverType: "axelar",
      recipientChain: toChainId(parsedLog.args.recipientChainId),
      digest: calculatedDigest,
    };
  } catch (err) {
    logger.error(`Error parsing axelar send transceiver message: ${err}`);
  }
};

const MAIN_TOPICS: Record<string, LogToNTTTransfer<NTTTransfer>> = {
  "0xe54e51e42099622516fa3b48e9733581c9dbdcb771cafb093f745a0532a35982": mapLogDataFromTransferSent,
};

const AXELAR_SEND_TRANSCEIVER_MESSAGE_TOPIC =
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a";
const WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC =
  "0x79376a0dc6cbfe6f6f8f89ad24c262a8c6233f8df181d3fe5abb2e2442e8c738";

const TRANSCEIVER_TOPICS: Record<string, LogToNTTTransfer<TransceiverLogData | undefined>> = {
  // Note: Keep Axelar topic first, as second topic is also present on Axelar transceiver contract
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
  } else {
    // If axelar topic is not present, then it is definitely wormhole transceiver
    const wormholeTransceiverLog = logs.find(
      (log) => log.topics[0] === WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC
    );

    if (!wormholeTransceiverLog) {
      logger.warn(`Couldn't find transceiver log in transaction logs`);
      return undefined;
    }

    const mapper = TRANSCEIVER_TOPICS[WORMHOLE_SEND_TRANSCEIVER_MESSAGE_TOPIC];
    return mapper(wormholeTransceiverLog, emitterChainId);
  }
};
