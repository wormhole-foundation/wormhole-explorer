import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { WormchainLog } from "../../../domain/entities/wormchain";
import winston from "winston";

const CHAIN_ID_WORMCHAIN = 3104;
const CORE_ADDRESS = "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j";

let logger: winston.Logger = winston.child({ module: "wormchainLogMessagePublishedMapper" });

export const wormchainLogMessagePublishedMapper = (
  log: WormchainLog
): LogFoundEvent<LogMessagePublished> | undefined => {
  const { coreContract, sequence, payload, emitter, nonce, hash } = transactionAttibutes(log);

  if (coreContract && sequence && payload && emitter && payload && nonce && hash) {
    logger.info(
      `[wormchain] Source event info: [tx: ${hash}][emitterChain: ${CHAIN_ID_WORMCHAIN}][sender: ${emitter} }}][sequence: ${sequence} ]`
    );

    return {
      name: "log-message-published",
      address: CORE_ADDRESS,
      chainId: 3104,
      txHash: hash,
      blockHeight: log.blockHeight,
      blockTime: log.timestamp,
      attributes: {
        sender: emitter,
        sequence: sequence,
        payload: payload,
        nonce: nonce,
        consistencyLevel: 0,
      },
    };
  }
};

function transactionAttibutes(log: WormchainLog): TransactionAttributes {
  let coreContract;
  let sequence;
  let payload;
  let emitter;
  let nonce;
  let hash;

  log.transactions?.forEach((tx) => {
    hash = tx.hash;

    tx.attributes.forEach((attr) => {
      const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
      const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

      switch (key) {
        case "message.sequence":
          console.log(key, value);
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
          if (value.toLocaleLowerCase() === CORE_ADDRESS.toLowerCase()) {
            coreContract = true;
          }
          break;
      }
    });
  });

  return {
    coreContract,
    sequence,
    payload,
    emitter,
    nonce,
    hash,
  };
}

type TransactionAttributes = {
  coreContract: boolean | undefined;
  sequence: number | undefined;
  payload: string | undefined;
  emitter: string | undefined;
  nonce: number | undefined;
  hash: string | undefined;
};
