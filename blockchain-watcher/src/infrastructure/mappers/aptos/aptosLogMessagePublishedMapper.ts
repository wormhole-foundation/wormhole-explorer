import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { TransactionsByVersion } from "../../repositories/aptos/AptosJsonRPCBlockRepository";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosLogMessagePublishedMapper" });

export const aptosLogMessagePublishedMapper = (
  tx: TransactionsByVersion
): LogFoundEvent<LogMessagePublished> | undefined => {
  if (!tx.blockTime) {
    throw new Error(`Block time is missing for tx ${tx.hash}`);
  }

  if (tx) {
    logger.info(
      `[aptos] Source event info: [tx: ${tx.hash}][emitterChain: 22][sender: ${tx.sender}}][sequence: ${tx.sequence}]`
    );

    return {
      name: "log-message-published",
      address: tx.address!,
      chainId: 22,
      txHash: tx.hash!,
      blockHeight: tx.blockHeight!,
      blockTime: tx.blockTime,
      attributes: {
        sender: tx.sender!,
        sequence: Number(tx.sequence!),
        payload: tx.payload!,
        nonce: Number(tx.nonce),
        consistencyLevel: tx.consistencyLevel!,
      },
    };
  }
};
