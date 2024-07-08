import { EvmLog, LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { BigNumber } from "ethers";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "evmLogMessagePublishedMapper" });

export const evmLogMessagePublishedMapper = (
  log: EvmLog,
  parsedArgs: ReadonlyArray<any>
): LogFoundEvent<LogMessagePublished> => {
  if (!log.blockTime) {
    throw new Error(`Block time is missing for log ${log.logIndex} in tx ${log.transactionHash}`);
  }

  const chainId = log.chainId;
  const txHash = log.transactionHash;
  const sender = parsedArgs[0];
  const sequence = (parsedArgs[1] as BigNumber).toNumber();

  logger.info(
    `[${log.chain}] Source event info: [tx: ${txHash}][VAA: ${chainId}/${sender}/${sequence}]`
  );

  return {
    name: "log-message-published",
    address: log.address,
    chainId: chainId,
    txHash: txHash,
    blockHeight: log.blockNumber,
    blockTime: log.blockTime,
    attributes: {
      sender: sender, // log.topics[1]
      sequence: sequence,
      nonce: parsedArgs[2],
      payload: parsedArgs[3],
      consistencyLevel: parsedArgs[4],
    },
  };
};
