import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { CosmosTransaction } from "../../../domain/entities/wormchain";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "wormchainLogMessagePublishedMapper" });

export const wormchainLogMessagePublishedMapper = (
  addresses: string[],
  transaction: CosmosTransaction
): LogFoundEvent<LogMessagePublished>[] | [] => {
  const transactionAttributesMapped = transactionAttributes(addresses, transaction);

  if (transactionAttributesMapped.length === 0) {
    return [];
  }

  const logMessages: LogFoundEvent<LogMessagePublished>[] = [];

  transactionAttributesMapped.forEach((attributes) => {
    logger.info(
      `[wormchain] Source event info: [tx: ${transaction.hash}][VAA: ${attributes.chainId}/${attributes.emitter}/${attributes.sequence}]`
    );

    logMessages.push({
      name: "log-message-published",
      address: attributes.coreContract!,
      chainId: attributes.chainId,
      txHash: attributes.hash!,
      blockHeight: transaction.blockHeight,
      blockTime: transaction.timestamp!,
      attributes: {
        sender: attributes.emitter!,
        sequence: attributes.sequence!,
        payload: attributes.payload!,
        nonce: attributes.nonce!,
        consistencyLevel: 0,
      },
    });
  });

  return logMessages;
};

function transactionAttributes(
  addresses: string[],
  transaction: CosmosTransaction
): TransactionAttributes[] {
  const transactionAttributes: TransactionAttributes[] = [];

  let coreContract: string | undefined;
  let sequence: number | undefined;
  let chainId: number | undefined;
  let payload: string | undefined;
  let emitter: string | undefined;
  let nonce: number | undefined;
  let hash: string | undefined;

  for (const attr of transaction.attributes) {
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
    hash = transaction.hash;
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
