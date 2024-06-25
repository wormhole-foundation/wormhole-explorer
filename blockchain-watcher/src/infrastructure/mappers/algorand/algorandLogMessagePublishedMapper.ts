import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
import winston from "winston";
import algosdk from "algosdk";

const CHAIN_ID_ALGORAND = 8;

let logger: winston.Logger = winston.child({ module: "algorandLogMessagePublishedMapper" });

export const algorandLogMessagePublishedMapper = (
  transaction: AlgorandTransaction
): LogFoundEvent<LogMessagePublished> | undefined => {
  if (!transaction.sender || !transaction.innerTxs || transaction.innerTxs.length === 0) {
    return undefined;
  }

  const innetTxwithLogs = transaction.innerTxs.find((tx) => tx.logs);

  if (!innetTxwithLogs || !innetTxwithLogs.logs || innetTxwithLogs.logs.length === 0) {
    return undefined;
  }

  // We use the sender address from innerTxs to build the emitterChain because the sender address
  // from the transaction is the bridge address (token bridge)
  const emitterChain = Buffer.from(
    algosdk.decodeAddress(innetTxwithLogs.sender).publicKey
  ).toString("hex");

  const sequence = Number(`0x${Buffer.from(innetTxwithLogs.logs[0], "base64").toString("hex")}`);

  logger.info(
    `[algorand] Source event info: [tx: ${transaction.hash}][${CHAIN_ID_ALGORAND}/${emitterChain}/${sequence}]`
  );

  return {
    name: "log-message-published",
    address: transaction.sender,
    chainId: CHAIN_ID_ALGORAND,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      sender: emitterChain,
      sequence: sequence,
      payload: transaction.payload,
      nonce: 0, // https://developer.algorand.org/docs/get-details/ethereum_to_algorand/#nonces-validity-windows-and-leases
      consistencyLevel: 0,
    },
  };
};
