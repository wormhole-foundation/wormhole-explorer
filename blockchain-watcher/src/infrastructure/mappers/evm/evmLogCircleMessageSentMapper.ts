import { EvmTransaction, LogFoundEvent, CircleMessageSent } from "../../../domain/entities";
import { deserializeCircleMessage } from "./helpers/circle";
import { encoding, circle } from "@wormhole-foundation/sdk-connect";
import { HandleEvmConfig } from "../../../domain/actions";
import { ethers } from "ethers";
import winston from "winston";

const WORMHOLE_TOPIC = "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2";
let logger: winston.Logger = winston.child({ module: "evmLogCircleMessageSentMapper" });

export const evmLogCircleMessageSentMapper = (
  transaction: EvmTransaction,
  cfg?: HandleEvmConfig
): LogFoundEvent<CircleMessageSent> | undefined => {
  const messageProtocol = mappedMessageProtocol(transaction.logs);
  const circleMessageSent = mappedCircleMessageSent(transaction.logs, cfg!);

  if (!circleMessageSent) {
    logger.warn(
      `[${transaction.chain}] Failed to parse circle message for [tx: ${transaction.hash}]`
    );
    return undefined;
  }

  logger.info(
    `[${transaction.chain}] Circle message sent event info: [tx: ${transaction.hash}] [protocol: ${circleMessageSent.protocol} - ${messageProtocol}]`
  );

  return {
    name: "circle-message-sent",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      ...circleMessageSent,
      txHash: transaction.hash,
    },
    tags: {
      destinationDomainMsg: circleMessageSent.destinationDomain,
      messageProtocolMsg: messageProtocol,
      sourceDomainMsg: circleMessageSent.sourceDomain,
      protocolMsg: circleMessageSent.protocol,
      senderMsg: circleMessageSent.sender,
    },
  };
};

const mappedCircleMessageSent = (
  logs: EvmTransactionLog[],
  cfg: HandleEvmConfig
): CircleMessageSent | undefined => {
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
  const bytes = encoding.hex.decode(parsedLog.args[0]);
  const [protocol, circleMessage] = deserializeCircleMessage(bytes);

  if (!circleMessage || protocol !== "cctp" || circleMessage.payload instanceof Uint8Array) {
    return undefined;
  }

  return {
    destinationCaller: circleMessage.destinationCaller.toString(),
    destinationDomain: toCirceChain(cfg.environment, circleMessage.destinationDomain),
    messageSender: circleMessage.payload.messageSender.toString(),
    mintRecipient: circleMessage.payload.mintRecipient.toString(),
    sourceDomain: toCirceChain(cfg.environment, circleMessage.sourceDomain),
    burnToken: circleMessage.payload.burnToken.toString(),
    recipient: circleMessage.recipient.toString(),
    sender: circleMessage.sender.toString(),
    amount: circleMessage.payload.amount,
    nonce: circleMessage.nonce,
    protocol,
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
