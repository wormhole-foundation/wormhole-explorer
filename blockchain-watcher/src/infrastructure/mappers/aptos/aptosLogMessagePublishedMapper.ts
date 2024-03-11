import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { AptosTransaction } from "../../../domain/entities/aptos";
import winston from "winston";

const CHAIN_ID_APTOS = 22;

let logger: winston.Logger = winston.child({ module: "aptosLogMessagePublishedMapper" });

export const aptosLogMessagePublishedMapper = (
  tx: AptosTransaction
): LogFoundEvent<LogMessagePublished> | undefined => {
  if (!tx.blockTime) {
    throw new Error(`[aptos] Block time is missing for tx ${tx.hash}`);
  }

  logger.info(
    `[aptos] Source event info: [tx: ${tx.hash}][emitterChain: ${CHAIN_ID_APTOS}][sender: ${tx.sender}}][sequence: ${tx.sequence}]`
  );

  return {
    name: "log-message-published",
    address: tx.address,
    chainId: CHAIN_ID_APTOS,
    txHash: tx.hash,
    blockHeight: tx.blockHeight,
    blockTime: tx.timestamp,
    attributes: {
      sender: tx.sender,
      sequence: Number(tx.sequence),
      payload: tx.payload,
      nonce: Number(tx.nonce),
      consistencyLevel: tx.consistencyLevel,
    },
  };
};
