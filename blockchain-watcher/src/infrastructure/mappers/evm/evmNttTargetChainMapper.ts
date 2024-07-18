import {
  EvmTransaction,
  EvmTransactionLog,
  EVMNTTManagerAttributes,
  TransactionFoundEvent,
} from "../../../domain/entities";
import winston from "winston";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import { deserializeNttMessageDigest, NTTTransfer } from "./helpers/ntt";
import { toChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";
import { LogToNTTTransfer, mapLogDataByTopic, mappedTxnStatus } from "./helpers/utils";

let logger: winston.Logger = winston.child({ module: "evmTargetChainNttMapper" });

export const evmTargetChainNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const first10Characters = transaction.input.slice(0, 10);
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    first10Characters,
    transaction.hash
  );
  const { type: protocolType, method: protocolMethod } = protocol;
  const nttTransferInfo = mapLogDataByTopic(TOPICS, transaction.logs, transaction.chainId);
  const txnStatus = mappedTxnStatus(transaction.status);

  if (!nttTransferInfo) {
    logger.warn(
      `[${transaction.chain}] Couldn't map ntt transfer: [hash: ${transaction.hash}][protocol: ${protocolType}/${protocolMethod}]`
    );
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
      from: transaction.from,
      to: transaction.to,
      status: txnStatus,
      blockNumber: transaction.blockNumber,
      methodsByAddress: protocolMethod,
      timestamp: transaction.timestamp,
      txHash: transaction.hash,
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      gasUsed: transaction.gasUsed,
      effectiveGasPrice: transaction.effectiveGasPrice,
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      protocol: protocolType,
      recipient: nttTransferInfo.recipient,
      amount: nttTransferInfo.amount,
      messageId: nttTransferInfo.messageId,
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
    },
  };
};

// TODO: Add common error handling in parsing logic

// Transfer redeemed (NTT Manager on destination chain)
/**
 * Two responsibilities:
 * 1. Push data point for completion of transfer
 * 2. One for time taken for e2e relay (transfer sent <> transfer redeemed)
 */
const mapLogDataFromTransferRedeemed: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi = "event TransferRedeemed(bytes32 indexed digest);";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const parsedDigest = deserializeNttMessageDigest(parsedLog.args.digest);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    eventName: "transfer-redeemed",
    amount: parsedDigest.payload.trimmedAmount.amount,
    recipient: parsedDigest.payload.recipientAddress.toNative(parsedDigest.payload.recipientChain),
    recipientChain: toChainId(parsedDigest.payload.recipientChain),
    emitterChain: toChainId(emitterChainId),
    messageId: Number(parsedDigest.id.toString()),
    sourceToken: parsedDigest.payload.sourceToken.toNative(emitterChain),
  };
};

/**
 * Two responsibilities:
 * 1. Push data point for a particular transceiver on receiving a relayed message
 * 2. One for time taken for a particular relay by the relayer (send transceiver <> receive relayed message)
 */
const mapLogDataFromReceivedRelayedMessage: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi =
    "event ReceivedRelayedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const parsedDigest = deserializeNttMessageDigest(parsedLog.args.digest);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    eventName: "received-relayed-message",
    amount: parsedDigest.payload.trimmedAmount.amount,
    recipient: parsedDigest.payload.recipientAddress.toNative(parsedDigest.payload.recipientChain),
    recipientChain: toChainId(parsedDigest.payload.recipientChain),
    emitterChain: toChainId(emitterChainId),
    messageId: Number(parsedDigest.id.toString()),
    sourceToken: parsedDigest.payload.sourceToken.toNative(emitterChain),
  };
};

/**
 * Two responsibilities:
 * 1. Push data point for a particular transceiver when it attests a message
 * 2. One for time taken by a particular tranceiver in total (send transceiver <> messageAttestedTo)
 */
const mapLogDataFromMessageAttestedTo: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi = "event MessageAttestedTo (bytes32 digest, address transceiver, uint8 index)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const parsedDigest = deserializeNttMessageDigest(parsedLog.args.digest);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    eventName: "message-attested-to",
    amount: parsedDigest.payload.trimmedAmount.amount,
    recipient: parsedDigest.payload.recipientAddress.toNative(parsedDigest.payload.recipientChain),
    recipientChain: toChainId(parsedDigest.payload.recipientChain),
    emitterChain: toChainId(emitterChainId),
    messageId: Number(parsedDigest.id.toString()),
    sourceToken: parsedDigest.payload.sourceToken.toNative(emitterChain),
  };
};

const TOPICS: Record<string, LogToNTTTransfer<NTTTransfer>> = {
  "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91":
    mapLogDataFromTransferRedeemed,
  "0xf557dbbb087662f52c815f6c7ee350628a37a51eae9608ff840d996b65f87475":
    mapLogDataFromReceivedRelayedMessage,
  "0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5":
    mapLogDataFromMessageAttestedTo,
};
