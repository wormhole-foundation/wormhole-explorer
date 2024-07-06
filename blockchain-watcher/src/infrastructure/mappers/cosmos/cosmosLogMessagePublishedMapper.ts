import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { CosmosTransaction } from "../../../domain/entities/Cosmos";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "cosmosLogMessagePublishedMapper" });

export const cosmosLogMessagePublishedMapper = (
  addresses: string[],
  transaction: CosmosTransaction
): LogFoundEvent<LogMessagePublished> | undefined => {
  const transactionAttributesMapped = transactionAttributes(addresses, transaction);

  if (!transactionAttributesMapped) {
    return undefined;
  }

  logger.info(
    `[${transaction.chain}] Source event info: [tx: ${transaction.hash}][VAA: ${transaction.chainId}/${transactionAttributesMapped.emitter}/${transactionAttributesMapped.sequence}]`
  );

  return {
    name: "log-message-published",
    address: transactionAttributesMapped.coreContract!,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: transaction.height,
    blockTime: transaction.timestamp!,
    attributes: {
      sender: transactionAttributesMapped.emitter!,
      sequence: transactionAttributesMapped.sequence!,
      payload: transactionAttributesMapped.payload!,
      nonce: transactionAttributesMapped.nonce!,
      consistencyLevel: 0,
      chain: transaction.chain,
    },
  };
};

function transactionAttributes(
  addresses: string[],
  transaction: CosmosTransaction
): TransactionAttributes | undefined {
  let transactionAttributes;

  transaction.events?.forEach((event) => {
    let coreContract: string | undefined;
    let sequence: number | undefined;
    let chainId: number | undefined;
    let payload: string | undefined;
    let emitter: string | undefined;
    let nonce: number | undefined;

    for (const attr of event.attributes) {
      const key = attr.key;
      const value = attr.value;

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
      transactionAttributes = {
        timestamp: transaction.timestamp,
        hash: transaction.hash!,
        coreContract,
        sequence,
        payload,
        chainId,
        emitter,
        nonce,
      };
    }
  });

  return transactionAttributes;
}

type TransactionAttributes = {
  coreContract: string | undefined;
  timestamp: string | undefined;
  sequence: number | undefined;
  payload: string | undefined;
  emitter: string | undefined;
  chainId: number | undefined;
  nonce: number | undefined;
  hash: string | undefined;
};
