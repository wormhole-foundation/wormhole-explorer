import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { WormchainBlockLogs } from "../../../domain/entities/wormchain";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "wormchainLogMessagePublishedMapper" });

export const wormchainRedeemedTransactionFoundMapper = (
  addresses: string[],
  log: WormchainBlockLogs
): LogFoundEvent<LogMessagePublished>[] | [] => {
  const transactionAttributesMapped = transactionAttributes(addresses, log);

  if (transactionAttributesMapped.length === 0) {
    return [];
  }

  const logMessages: LogFoundEvent<LogMessagePublished>[] = [];

  transactionAttributesMapped.forEach((tx) => {
    logger.info(
      `[wormchain] Source event info: [tx: ${tx.hash}][emitterChain: ${tx.chainId}][sender: ${tx.emitter}][sequence: ${tx.sequence}]`
    );

    logMessages.push({
      name: "log-message-published",
      address: tx.coreContract!,
      chainId: tx.chainId,
      txHash: tx.hash!,
      blockHeight: log.blockHeight,
      blockTime: log.timestamp,
      attributes: {
        sender: tx.emitter!,
        sequence: tx.sequence!,
        payload: tx.payload!,
        nonce: tx.nonce!,
        consistencyLevel: 0,
      },
    });
  });

  return logMessages;
};

function transactionAttributes(
  addresses: string[],
  log: WormchainBlockLogs
): TransactionAttributes[] {
  const transactionAttributes: TransactionAttributes[] = [];

  log.transactions?.forEach((tx) => {
    let srcChannel: string | undefined;
    let dstChannel: string | undefined;
    let timestamp: string | undefined;
    let receiver: string | undefined;
    let sequence: number | undefined;
    let sender: string | undefined;

    for (const attr of tx.attributes) {
      const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
      const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

      switch (key) {
        case "packet_data":
          const packetData = JSON.parse(value) as PacketData;
          sequence = Number(value);
          break;
      }
    }

    if (srcChannel) {
      transactionAttributes.push({
        srcChannel,
        dstChannel,
        timestamp,
        receiver,
        sequence,
        sender,
      });
    }
  });

  return transactionAttributes;
}
