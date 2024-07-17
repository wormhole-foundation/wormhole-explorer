import {
  EvmTransaction,
  EvmTransactionLog,
  EVMNTTManagerAttributes,
  TransactionFoundEvent,
} from "../../../domain/entities";
import winston from "winston";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import {
  deserializeNttMessageDigest,
  deserializeWormholeTransceiverMessage,
  NTTTransfer,
} from "./helpers/ntt";
import { toChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";
import { UniversalAddress } from "@wormhole-foundation/sdk-definitions";
import { LogToNTTTransfer, mapLogDataByTopic, mappedTxnStatus } from "./helpers/utils";

let logger: winston.Logger = winston.child({ module: "evmSourceChainNttMapper" });

export const evmSourceChainNttMapper = (
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
      emitterChain: transaction.chainId,
    },
  };
};

// Transfer sent (NTT Manager on source chain)
const mapLogDataFromTransferSent: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi =
    "event TransferSent(bytes32 recipient, uint256 amount, uint256 fee, uint16 recipientChain, uint64 msgSequence)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "transfer-sent",
    recipient: new UniversalAddress(parsedLog.args.recipient).toNative(
      parsedLog.args.recipientChain
    ),
    amount: BigInt(parsedLog.args.amount),
    fee: BigInt(parsedLog.args.fee),
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: Number(parsedLog.args.msgSequence),
    sourceToken: undefined,
  };
};

const mapLogDataWormholeSendTransceiverMessage: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi = "event SendTransceiverMessage(uint16 recipientChain, tuple message)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const message = deserializeWormholeTransceiverMessage(parsedLog.args.message);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    eventName: "send-transceiver-message",
    transceiverType: "wormhole",
    recipient: message.nttManagerPayload.payload.recipientAddress.toNative(
      parsedLog.args.recipientChain
    ),
    amount: message.nttManagerPayload.payload.trimmedAmount.amount,
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: Number(message.nttManagerPayload.id.toString()),
    sourceToken: message.nttManagerPayload.payload.sourceToken.toNative(emitterChain),
  };
};

const mapLogDataAxelarSendTransceiverMessage: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog,
  emitterChainId: number
): NTTTransfer => {
  const abi =
    "event SendTransceiverMessage(index_topic_1 uint16 recipientChainId, bytes nttManagerMessage, index_topic_2 bytes32 recipientNttManagerAddress, index_topic_3 bytes32 refundAddress)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const message = deserializeNttMessageDigest(parsedLog.args.nttManagerMessage);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    eventName: "send-transceiver-message",
    transceiverType: "axelar",
    recipient: message.payload.recipientAddress.toNative(parsedLog.args.recipientChain),
    amount: message.payload.trimmedAmount.amount,
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: Number(message.id.toString()),
    sourceToken: message.payload.sourceToken.toNative(emitterChain),
  };
};

const TOPICS: Record<string, LogToNTTTransfer<NTTTransfer>> = {
  "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf": mapLogDataFromTransferSent,
  "0x53b3e029c5ead7bffc739118953883859d30b1aaa086e0dca4d0a1c99cd9c3f5":
    mapLogDataWormholeSendTransceiverMessage,
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a":
    mapLogDataAxelarSendTransceiverMessage,
};
