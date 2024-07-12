import {
  EvmTransaction,
  EvmTransactionLog,
  EVMTransferSentAttributes,
  TransactionFoundEvent,
} from "../../../domain/entities";
import winston from "winston";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import { deserializeNttMessageDigest } from "./helpers/ntt";
import { ChainId, toChainId, isChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";

let logger: winston.Logger = winston.child({ module: "evmTransferSentNttMapper" });

export const evmTransferSentNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMTransferSentAttributes> => {
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

const mapLogDataFromTransferSent = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi =
    "event TransferSent(bytes32 recipient, uint256 amount, uint256 fee, uint16 recipientChain, uint64 msgSequence)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    recipient: undefined,
    amount: BigInt(parsedLog.args.amount),
    fee: BigInt(parsedLog.args.fee),
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: Number(parsedLog.args.msgSequence),
    sourceToken: undefined,
  };
};

const mapLogDataFromTransferRedeemed = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi = "event TransferRedeemed(bytes32 indexed digest);";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const parsedDigest = deserializeNttMessageDigest(parsedLog.args.digest);
  const emitterChain = chainIdToChain(toChainId(emitterChainId));

  return {
    amount: undefined,
    fee: undefined,
    recipient: parsedDigest.payload.recipientAddress.toNative(parsedDigest.payload.recipientChain),
    recipientChain: toChainId(parsedDigest.payload.recipientChain),
    messageId: parsedDigest.id.toString(),
    sourceToken: parsedDigest.payload.sourceToken.toNative(emitterChain),
  };
};

const TOPICS: Record<string, unknown> = {
  "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf": mapLogDataFromTransferSent,
  "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91":
    mapLogDataFromTransferRedeemed,
};
