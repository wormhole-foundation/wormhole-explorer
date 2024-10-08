import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { decodeNttTransferSent, EVMNTTManagerAttributes, NTTTransfer } from "./helpers/ntt";
import { toChainId } from "@wormhole-foundation/sdk-base";
import {
  isTopicPresentInLogs,
  LogMapperFn,
  mapLogDataByTopic,
  mapTxnStatus,
} from "./helpers/utils";
import { evmNttWormholeTransceiverMapper } from "./evmNttWormholeTransceiverMapper";

let logger: winston.Logger = winston.child({ module: "evmNttTransferSentMapper" });

export const evmNttTransferSentMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const emitterChainId = toChainId(transaction.chainId);

  // Process further only if transaction logs contain NTT_MANAGER_TOPICS
  if (!isTopicPresentInLogs(NTT_MANAGER_TOPICS, transaction.logs)) {
    return undefined;
  }

  const transceiverInfo = evmNttWormholeTransceiverMapper(transaction);

  if (!transceiverInfo) {
    logger.warn(
      `[${transaction.chain}] Couldn't map transceiver type: [hash: ${transaction.hash}]`
    );
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
      blockNumber: BigInt(transaction.blockNumber),
      timestamp: transaction.timestamp,
      txHash: transaction.hash,
      gas: BigInt(transaction.gas),
      gasPrice: BigInt(transaction.gasPrice),
      gasUsed: BigInt(transaction.gasUsed),
      effectiveGasPrice: BigInt(transaction.effectiveGasPrice),
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      protocol: "NTT",
      recipient: nttTransferInfo.recipient,
      amount: transceiverInfo?.attributes?.amount,
      // We use digest as an unique identifier for the NTT transfer events across source and target chains
      digest: transceiverInfo.attributes.digest,
      ...(nttTransferInfo?.fee && {
        fee: nttTransferInfo?.fee,
      }),
    },
    tags: {
      recipientChain: nttTransferInfo.recipientChain,
      emitterChain: nttTransferInfo.emitterChain,
      sourceToken: transceiverInfo?.tags?.sourceToken,
    },
  };
};

export const mapLogDataFromTransferSent: LogMapperFn<Omit<NTTTransfer, "digest">> = (
  log: EvmTransactionLog,
  emitterChainId: number
): Omit<NTTTransfer, "digest"> => {
  const parsedLog = decodeNttTransferSent(log.data);
  const recipientChainId = toChainId(parsedLog.recipientChain);

  return {
    eventName: "ntt-transfer-sent",
    recipient: parsedLog.recipient,
    amount: BigInt(parsedLog.amount),
    fee: BigInt(parsedLog.fee),
    recipientChain: recipientChainId,
    emitterChain: toChainId(emitterChainId),
  };
};

const NTT_MANAGER_TOPICS: Record<string, LogMapperFn<Omit<NTTTransfer, "digest">>> = {
  "0xe54e51e42099622516fa3b48e9733581c9dbdcb771cafb093f745a0532a35982": mapLogDataFromTransferSent,
};
