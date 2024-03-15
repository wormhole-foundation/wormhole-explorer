import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { AptosTransaction } from "../../../domain/entities/aptos";
import winston from "winston";

const REDEEM_EVENT_TAIL = "::state::WormholeMessage";
const CHAIN_ID_APTOS = 22;

let logger: winston.Logger = winston.child({ module: "aptosLogMessagePublishedMapper" });

export const aptosLogMessagePublishedMapper = (
  transaction: AptosTransaction
): LogFoundEvent<LogMessagePublished> | undefined => {
  const wormholeEvent = transaction.events.find((tx: any) => tx.type.includes(REDEEM_EVENT_TAIL));
  const wormholeData = wormholeEvent.data;

  const address = transaction.payload.function.split("::")[0];

  logger.info(
    `[aptos] Source event info: [tx: ${transaction.hash}][emitterChain: ${CHAIN_ID_APTOS}][sender: ${wormholeData.sender}}][sequence: ${wormholeData.sequence}]`
  );

  return {
    name: "log-message-published",
    address: address,
    chainId: CHAIN_ID_APTOS,
    txHash: transaction.hash,
    blockHeight: transaction.blockHeight,
    blockTime: wormholeData.timestamp,
    attributes: {
      sender: wormholeData.sender,
      sequence: Number(wormholeData.sequence),
      payload: wormholeData.payload,
      nonce: Number(wormholeData.nonce),
      consistencyLevel: wormholeData.consistencyLevel,
    },
  };
};
