import {
  EvmTransaction,
  EvmTransactionLog,
  EVMNTTManagerAttributes,
  TransactionFoundEvent,
} from "../../../domain/entities";
import winston from "winston";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import { deserializeNttMessageDigest, deserializeWormholeTransceiverMessage } from "./helpers/ntt";
import { ChainId, toChainId, isChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";
import { UniversalAddress } from "@wormhole-foundation/sdk-definitions";

let logger: winston.Logger = winston.child({ module: "evmSourceChainNttMapper" });

export const evmSourceChainNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> => {
  const first10Characters = transaction.input.slice(0, 10);
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    first10Characters,
    transaction.hash
  );
  const { type: protocolType, method: protocolMethod } = protocol;

  // get attributes
  // get tags -> emitterChain, receipientChain
  // for transceiver: transceiverType(axelar, wormhole)
};

const mapLogDataByTopic = (emitterChainId: number, logs: EvmTransactionLog[]) => {
  const filterLogs = logs.filter((log) => {
    return TOPICS[log.topics[0]];
  });

  if (!filterLogs) return undefined;
};

// Transfer sent (NTT Manager on source chain)
const mapLogDataFromTransferSent = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi =
    "event TransferSent(bytes32 recipient, uint256 amount, uint256 fee, uint16 recipientChain, uint64 msgSequence)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
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

const mapLogDataWormholeSendTransceiverMessage = (
  emitterChainId: number,
  log: EvmTransactionLog
) => {
  const abi = "event SendTransceiverMessage(uint16 recipientChain, tuple message)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const message = deserializeWormholeTransceiverMessage(parsedLog.args.message);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    recipient: message.nttManagerPayload.payload.recipientAddress.toNative(
      parsedLog.args.recipientChain
    ),
    amount: message.nttManagerPayload.payload.trimmedAmount.amount,
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: message.nttManagerPayload.id.toString(),
    sourceToken: message.nttManagerPayload.payload.sourceToken.toNative(emitterChain),
  };
};

const mapLogDataAxelarSendTransceiverMessage = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi =
    "event SendTransceiverMessage(index_topic_1 uint16 recipientChainId, bytes nttManagerMessage, index_topic_2 bytes32 recipientNttManagerAddress, index_topic_3 bytes32 refundAddress)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const message = deserializeNttMessageDigest(parsedLog.args.nttManagerMessage);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    recipient: message.payload.recipientAddress.toNative(parsedLog.args.recipientChain),
    amount: message.payload.trimmedAmount.amount,
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: message.id.toString(),
    sourceToken: message.payload.sourceToken.toNative(emitterChain),
  };
};

const TOPICS: Record<string, unknown> = {
  "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf": mapLogDataFromTransferSent,
  "0x53b3e029c5ead7bffc739118953883859d30b1aaa086e0dca4d0a1c99cd9c3f5":
    mapLogDataWormholeSendTransceiverMessage,
  "0xcdba4baae54ffe4453599128e176cfa8a3190fff44e9f60a444875db7fb0572a":
    mapLogDataAxelarSendTransceiverMessage,
};
