import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { CosmosRedeem } from "../../../domain/entities/wormchain";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "cosmosLogMessagePublishedMapper" });

export const cosmosLogMessagePublishedMapper = (
  addresses: string[],
  tx: CosmosRedeem
): LogFoundEvent<LogMessagePublished> | undefined => {
  const transactionAttributesMapped = transactionAttributes(addresses, tx);

  if (!transactionAttributesMapped) {
    return undefined;
  }

  logger.info(
    `[${tx.chain}] Source event info: [tx: ${tx.hash}][VAA: ${tx.chainId}/${transactionAttributesMapped.emitter}/${transactionAttributesMapped.sequence}]`
  );

  return {
    name: "log-message-published",
    address: transactionAttributesMapped.coreContract!,
    chainId: tx.chainId,
    txHash: tx.hash,
    blockHeight: tx.height,
    blockTime: Number(tx.timestamp),
    attributes: {
      sender: transactionAttributesMapped.emitter!,
      sequence: transactionAttributesMapped.sequence!,
      payload: transactionAttributesMapped.payload!,
      nonce: transactionAttributesMapped.nonce!,
      consistencyLevel: 0,
      chain: tx.chain,
    },
  };
};

function transactionAttributes(
  addresses: string[],
  tx: CosmosRedeem
): TransactionAttributes | undefined {
  let transactionAttributes;

  tx.events?.forEach((event) => {
    let coreContract: string | undefined;
    let sequence: number | undefined;
    let chainId: number | undefined;
    let payload: string | undefined;
    let emitter: string | undefined;
    let nonce: number | undefined;
    let hash: string | undefined;

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
      hash = tx.hash!;
      transactionAttributes = {
        timestamp: tx.timestamp,
        coreContract,
        sequence,
        payload,
        chainId,
        emitter,
        nonce,
        hash,
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
  chainId: number;
  nonce: number | undefined;
  hash: string | undefined;
};
