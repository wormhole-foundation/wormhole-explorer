import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { WormchainLog } from "../../../domain/entities/wormchain";
import winston from "winston";

const CHAIN_ID_WORMCHAIN = 3104;
const CORE_ADDRESS = "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j";

let logger: winston.Logger = winston.child({ module: "wormchainLogMessagePublishedMapper" });

export const wormchainLogMessagePublishedMapper = (
  log: WormchainLog
): LogFoundEvent<LogMessagePublished>[] | [] => {
  const transactionAttributesMapped = transactionAttributes(log);

  if (transactionAttributesMapped.length === 0) {
    return [];
  }

  const logMessages: LogFoundEvent<LogMessagePublished>[] = [];

  transactionAttributesMapped.forEach((tx) => {
    logger.info(
      `[wormchain] Source event info: [tx: ${tx.hash}][emitterChain: ${CHAIN_ID_WORMCHAIN}][sender: ${tx.emitter}][sequence: ${tx.sequence}]`
    );

    logMessages.push({
      name: "log-message-published",
      address: CORE_ADDRESS,
      chainId: CHAIN_ID_WORMCHAIN,
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

function transactionAttributes(log: WormchainLog): TransactionAttributes[] {
  const transactionAttributes: TransactionAttributes[] = [];

  log.transactions?.forEach((tx) => {
    let coreContract = false;
    let sequence: number | undefined;
    let payload: string | undefined;
    let emitter: string | undefined;
    let nonce: number | undefined;
    let hash: string | undefined;

    for (const attr of tx.attributes) {
      const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
      const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

      switch (key) {
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
          if (value.toLowerCase() === CORE_ADDRESS.toLowerCase()) {
            coreContract = true;
          }
          break;
      }
    }

    if (coreContract && sequence && payload && emitter && nonce) {
      hash = tx.hash;
      transactionAttributes.push({ coreContract, sequence, payload, emitter, nonce, hash });
    }
  });

  return transactionAttributes;
}

type TransactionAttributes = {
  coreContract: boolean | undefined;
  sequence: number | undefined;
  payload: string | undefined;
  emitter: string | undefined;
  nonce: number | undefined;
  hash: string | undefined;
};
