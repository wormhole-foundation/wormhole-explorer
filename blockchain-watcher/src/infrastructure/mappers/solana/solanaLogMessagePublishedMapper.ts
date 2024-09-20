import { solana, LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { normalizeCompileInstruction } from "./utils";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { configuration } from "../../config";
import winston from "winston";

const connection = new Connection(configuration.chains.solana.rpcs[0]); // TODO: should be better to inject this to improve testability

let logger: winston.Logger;
logger = winston.child({ module: "solanaLogMessagePublishedMapper" });

export const solanaLogMessagePublishedMapper = async (
  tx: solana.Transaction,
  { programId, commitment }: { programId: string; commitment?: Commitment }
): Promise<LogFoundEvent<LogMessagePublished>[]> => {
  if (!tx || !tx.blockTime) {
    throw new Error(
      `Block time is missing for tx ${tx?.transaction?.signatures} in slot ${tx?.slot}`
    );
  }

  if (tx.meta?.err) {
    logger.info(
      `[solana] Ignoring tx ${tx.transaction.signatures[0]} because it failed: ${JSON.stringify(
        tx.meta.err
      )}`
    );
    return [];
  }

  const message = tx.transaction.message;
  const accountKeys = message.accountKeys;
  const programIdIndex = accountKeys.findIndex((i) => i === programId);
  const instructions = message.compiledInstructions;
  const innerInstructions =
    tx.meta?.innerInstructions?.flatMap((i) => i.instructions.map(normalizeCompileInstruction)) ||
    [];

  const whInstructions = innerInstructions
    .concat(instructions)
    .filter((i) => i.programIdIndex === programIdIndex);

  const results: LogFoundEvent<LogMessagePublished>[] = [];
  for (const instruction of whInstructions) {
    // skip if not postMessage instruction
    const instructionId = instruction.data;
    if (instructionId[0] !== 0x01) {
      continue;
    }

    const accountId = accountKeys[instruction.accountKeyIndexes[1]];

    const txHash = tx.transaction.signatures[0];
    const vaaInformation = await getVaaInformation(accountId, commitment);
    if (!vaaInformation) {
      logger.warn(
        `[solana}] Cannot mapper vaa information: [hash: ${txHash}][account: ${accountId}]`
      );
      return [];
    }

    const { sequence, emitterAddress, emitterChain, nonce, payload, consistencyLevel } =
      vaaInformation || {};

    logger.info(
      `[solana] Source event info: [hash: ${txHash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
    );

    results.push({
      name: "log-message-published",
      address: programId,
      chainId: emitterChain,
      txHash: txHash,
      blockHeight: BigInt(tx.slot.toString()),
      blockTime: tx.blockTime,
      attributes: {
        sender: emitterAddress,
        sequence: Number(sequence),
        payload: payload.toString("hex"),
        nonce,
        consistencyLevel,
      },
    });
  }

  return results;
};

const getVaaInformation = async (
  accountAddress: string,
  commitment?: Commitment
): Promise<VaaInformation | undefined> => {
  try {
    const { message } = await getPostedMessage(connection, accountAddress, commitment);
    return {
      emitterChain: message.emitterChain,
      emitterAddress: message.emitterAddress.toString("hex"),
      sequence: Number(message.sequence),
      nonce: message.nonce,
      payload: message.payload,
      consistencyLevel: message.consistencyLevel,
    };
  } catch (e) {
    // If we can't get the message, we can't process the transaction
    return undefined;
  }
};

type VaaInformation = {
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  nonce: number;
  payload: Buffer;
  consistencyLevel: number;
};
