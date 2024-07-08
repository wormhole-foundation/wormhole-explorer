import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { WormchainBlockLogs } from "../../../domain/entities/wormchain";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "wormchainLogMessagePublishedMapper" });

export const wormchainLogMessagePublishedMapper = (
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
      `[wormchain] Source event info: [tx: ${tx.hash}][VAA: ${tx.chainId}/${tx.emitter}/${tx.sequence}]`
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
    let coreContract: string | undefined;
    let sequence: number | undefined;
    let chainId: number | undefined;
    let payload: string | undefined;
    let emitter: string | undefined;
    let nonce: number | undefined;
    let hash: string | undefined;

    for (const attr of tx.attributes) {
      const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
      const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

      switch (key) {
        case "message.chain_id":
          chainId = Number(value);
          break;
        case "message.sequence":
          sequence = Number(value);
          break;
        case "message.message":
          payload = value;
          break;
        case "message.sender":
          emitter = value;
          break;
        case "message.nonce":
          nonce = Number(value);
          break;
        case "_contract_address":
        case "contract_address":
          if (addresses.includes(value.toLowerCase())) {
            coreContract = value.toLowerCase();
          }
          break;
      }
    }

    if (coreContract && chainId && sequence && payload && emitter && nonce != undefined) {
      hash = tx.hash;
      transactionAttributes.push({
        coreContract,
        sequence,
        payload,
        chainId,
        emitter,
        nonce,
        hash,
      });
    }
  });

  return transactionAttributes;
}

type TransactionAttributes = {
  coreContract: string | undefined;
  sequence: number | undefined;
  payload: string | undefined;
  emitter: string | undefined;
  chainId: number;
  nonce: number | undefined;
  hash: string | undefined;
};
