import { EvmTransaction, LogFoundEvent, MessageSent } from "../../../domain/entities";
import { encoding, circle } from "@wormhole-foundation/sdk-connect";
import { HandleEvmConfig } from "../../../domain/actions";
import { CircleBridge } from "@wormhole-foundation/sdk-definitions";
import { ethers } from "ethers";
import winston from "winston";

const WORMHOLE_TOPIC = "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2";
let logger: winston.Logger = winston.child({ module: "evmLogCircleMessageSentMapper" });

export const evmLogCircleMessageSentMapper = (
  transaction: EvmTransaction,
  cfg?: HandleEvmConfig
): LogFoundEvent<MessageSent> | undefined => {
  const messageProtocol = mappedMessageProtocol(transaction.logs);
  const messageSent = mappedMessageSent(transaction.logs, cfg!);

  if (!messageSent) {
    logger.warn(`[${transaction.chain}] No message sent event found [tx: ${transaction.hash}]`);
    return undefined;
  }

  logger.info(`[${transaction.chain}] Message sent event info: [tx: ${transaction.hash}]`);

  return {
    name: "circle-message-sent",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      ...messageSent,
    },
    tags: {
      destinationDomain: messageSent.destinationDomain,
      messageProtocol: messageProtocol,
      sourceDomain: messageSent.sourceDomain,
      protocol: messageSent.protocol,
      sender: messageSent.sender,
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

  const circleBody = deserializedMsg[0];
  return {
    destinationCaller: circleBody.destinationCaller.toString(),
    destinationDomain: toCirceChain(cfg.environment, circleBody.destinationDomain),
    messageSender: circleBody.payload.messageSender.toString(),
    mintRecipient: circleBody.payload.mintRecipient.toString(),
    sourceDomain: toCirceChain(cfg.environment, circleBody.sourceDomain),
    burnToken: circleBody.payload.burnToken.toString(),
    recipient: circleBody.recipient.toString(),
    protocol: "cctp",
    sender: circleBody.sender.toString(),
    amount: circleBody.payload.amount,
    nonce: circleBody.nonce,
  };
};

const mappedMessageProtocol = (logs: EvmTransactionLog[]): string => {
  return logs.some((log) => log.topics[0] === WORMHOLE_TOPIC)
    ? MessageProtocol.Wormhole
    : MessageProtocol.None;
};

const toCirceChain = (env: string, domain: number) => {
  // Remove this when the SDK is updated to accept Noble as a domain with 4 value
  // @wormhole-foundation/sdk-base/dist/cjs/constants/circle.d.ts
  const environment = env === "mainnet" ? "Mainnet" : "Testnet";
  return domain === 4 ? "Noble" : circle.toCircleChain(environment, domain);
};

const EVENT_TOPICS: Record<string, LogToVaaMapper> = {
  "0x8c5261668696ce22758910d05bab8f186d6eb247ceac2af2e82c7dc17669b036": mapCircleBodyFromTopics, // CCTP MessageSent (circle bridge)
};

enum MessageProtocol {
  Wormhole = "wormhole",
  None = "",
}

type LogToVaaMapper = (log: EvmTransactionLog, cfg: HandleEvmConfig) => any | undefined;

type EvmTransactionLog = { address: string; topics: string[]; data: string };
