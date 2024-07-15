import {
  EvmTransaction,
  EvmTransactionLog,
  EVMNTTManagerAttributes,
  TransactionFoundEvent,
} from "../../../domain/entities";
import winston from "winston";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import { deserializeNttMessageDigest } from "./helpers/ntt";
import { ChainId, toChainId, isChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";

let logger: winston.Logger = winston.child({ module: "evmNttMapper" });

export const evmNttMapper = (
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
    recipient: undefined,
    amount: BigInt(parsedLog.args.amount),
    fee: BigInt(parsedLog.args.fee),
    recipientChain: toChainId(parsedLog.args.recipientChain),
    messageId: Number(parsedLog.args.msgSequence),
    sourceToken: undefined,
  };
};

// Transfer redeemed (NTT Manager on destination chain)
/**
 * Two responsibilities:
 * 1. Push data point for completion of transfer
 * 2. One for time taken for e2e relay
 */
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

const mapLogDataFromReceivedRelayedMessage = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi =
    "event ReceivedRelayedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress)";
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

const mapLogDataFromMessageAttestedTo = (emitterChainId: number, log: EvmTransactionLog) => {
  const abi = "";
};

const TOPICS: Record<string, unknown> = {
  "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf": mapLogDataFromTransferSent,
  "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91":
    mapLogDataFromTransferRedeemed,
  "0xf557dbbb087662f52c815f6c7ee350628a37a51eae9608ff840d996b65f87475":
    mapLogDataFromReceivedRelayedMessage,
  "0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5":
    mapLogDataFromMessageAttestedTo,
};

// persist blockHash, blockTime for each cron job

// NTT manager -> WH Transceiver -> Relayer -> WH Transceiver -> NTT manager -> TransferRedeemed

/*

Approach 1:

Index on source chain & persit it in some DB (mongoDB) -> MessageId 

Index target chain events, digest 


Approach 2:

Index events on target chain and scan a particular event in source chain 


*/
