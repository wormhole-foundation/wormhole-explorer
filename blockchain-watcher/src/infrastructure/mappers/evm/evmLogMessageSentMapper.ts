import { EvmTransaction, LogFoundEvent, MessageSent } from "../../../domain/entities";
import { encoding, circle } from "@wormhole-foundation/sdk-connect";
import { HandleEvmConfig } from "../../../domain/actions";
import { CircleBridge } from "@wormhole-foundation/sdk-definitions";
import { ethers } from "ethers";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "evmLogMessageSentMapper" });

export const evmLogMessageSentMapper = (
  transaction: EvmTransaction,
  cfg?: HandleEvmConfig
): LogFoundEvent<MessageSent> | undefined => {
  const messageSent = mappedMessageSent(transaction.logs, cfg!);

  if (!messageSent) {
    logger.warn(`[${transaction.chain}] No message sent event found [tx: ${transaction.hash}]`);
    return undefined;
  }

  logger.info(`[${transaction.chain}] Message sent event info: [tx: ${transaction.hash}]`);

  return {
    name: "message-sent",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      ...messageSent,
    },
  };
};

const mappedMessageSent = (
  logs: EvmTransactionLog[],
  cfg: HandleEvmConfig
): MessageSent | undefined => {
  const filterLogs = logs.filter((log) => {
    return EVENT_TOPICS[log.topics[0]];
  });

  if (!filterLogs) return undefined;

  for (const log of filterLogs) {
    const mapper = EVENT_TOPICS[log.topics[0]];
    const bodyMessage = mapper(log, cfg);

    if (bodyMessage) {
      return bodyMessage;
    }
  }
};

const mapCircleBodyFromTopics: LogToVaaMapper = (log: EvmTransactionLog, cfg: HandleEvmConfig) => {
  if (!log.topics[0]) {
    return undefined;
  }

  const iface = new ethers.utils.Interface([cfg.abi]);
  const parsedLog = iface.parseLog(log);
  const deserializedMsg = CircleBridge.deserialize(encoding.hex.decode(parsedLog.args[0]));

  if (!deserializedMsg || !deserializedMsg[0]) {
    return undefined;
  }

  const environment = cfg.environment === "mainnet" ? "Mainnet" : "Testnet";
  const circleBody = deserializedMsg[0];

  // Filter out messages that are not from or to the circle chain
  if (circleBody.destinationDomain === 4 || circleBody.sourceDomain === 4) {
    return undefined;
  }

  return {
    destinationAddress: circleBody.destinationCaller.toString(),
    destinationDomain: circle.toCircleChain(environment, circleBody.destinationDomain),
    recipientAddress: circleBody.recipient.toString(),
    senderAddress: circleBody.sender.toString(),
    messageSender: circleBody.payload.messageSender.toString(),
    mintRecipient: circleBody.payload.mintRecipient.toString(),
    sourceDomain: circle.toCircleChain(environment, circleBody.sourceDomain),
    burnToken: circleBody.payload.burnToken.toString(),
    amount: circleBody.payload.amount,
    nonce: circleBody.nonce,
  };
};

const EVENT_TOPICS: Record<string, LogToVaaMapper> = {
  "0x8c5261668696ce22758910d05bab8f186d6eb247ceac2af2e82c7dc17669b036": mapCircleBodyFromTopics, // CCTP MessageSent (circle bridge)
};

type LogToVaaMapper = (log: EvmTransactionLog, cfg: HandleEvmConfig) => any | undefined;

type EvmTransactionLog = { address: string; topics: string[]; data: string };
