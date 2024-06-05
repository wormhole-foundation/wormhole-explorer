import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { CHAIN_ID_SUI } from "@certusone/wormhole-sdk";
import { SuiEvent } from "@mysten/sui.js/client";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "suiLogMessagePublishedMapper" });

const SOURCE_EVENT_TAIL = "::publish_message::WormholeMessage";

export const suiLogMessagePublishedMapper = (
  receipt: SuiTransactionBlockReceipt
): LogFoundEvent<LogMessagePublished> | undefined => {
  const { events } = receipt;

  const event = events.find((e) => e.type.endsWith(SOURCE_EVENT_TAIL));
  if (!event) return undefined;

  const logMessage = extractEventInfo(event);
  if (!logMessage) return undefined;
  const { nonce, sender, sequence, payload, consistencyLevel } = logMessage;

  if (sender && sequence) {
    logger.info(
      `[sui] Source event info: [digest: ${receipt.digest}][VAA: ${CHAIN_ID_SUI}/${sender}/${sequence}]`
    );

    return {
      name: "log-message-published",
      address: event.packageId,
      blockHeight: BigInt(receipt.checkpoint || 0),
      blockTime: Math.floor(Number(receipt.timestampMs) / 1000), // convert to seconds
      chainId: CHAIN_ID_SUI,
      txHash: receipt.digest,
      attributes: {
        sender,
        sequence,
        payload,
        nonce,
        consistencyLevel,
      },
    };
  }
};

function extractEventInfo(event: SuiEvent): LogMessagePublished | undefined {
  const json = event.parsedJson as SuiSourceEvent;

  return {
    nonce: json.nonce,
    sender: json.sender,
    sequence: Number(json.sequence),
    payload: Buffer.from(json.payload).toString("hex"),
    consistencyLevel: json.consistency_level,
  };
}

interface SuiSourceEvent {
  nonce: number;
  sender: string;
  sequence: string;
  payload: number[];
  consistency_level: number;
}
