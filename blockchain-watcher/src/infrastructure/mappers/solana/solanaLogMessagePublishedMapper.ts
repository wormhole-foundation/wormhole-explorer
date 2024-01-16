import { decode } from "bs58";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { solana, LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { CompiledInstruction, MessageCompiledInstruction } from "../../../domain/entities/solana";
import { configuration } from "../../config";

const connection = new Connection(configuration.chains.solana.rpcs[0]); // TODO: should be better to inject this to improve testability

export const solanaLogMessagePublishedMapper = async (
  tx: solana.Transaction,
  { programId, commitment }: { programId: string; commitment?: Commitment }
): Promise<LogFoundEvent<LogMessagePublished>[]> => {
  if (!tx || !tx.blockTime) {
    throw new Error(
      `Block time is missing for tx ${tx?.transaction?.signatures} in slot ${tx?.slot}`
    );
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
    const { message } = await getPostedMessage(connection, accountId, commitment);
    const {
      sequence,
      emitterAddress,
      emitterChain,
      submissionTime: timestamp,
      nonce,
      payload,
      consistencyLevel,
    } = message || {};

    results.push({
      name: "log-message-published",
      address: programId,
      chainId: emitterChain,
      txHash: tx.transaction.signatures[0],
      blockHeight: BigInt(tx.slot.toString()),
      blockTime: tx.blockTime,
      attributes: {
        sender: emitterAddress.toString("hex"),
        sequence: Number(sequence),
        payload: payload.toString("hex"),
        nonce,
        consistencyLevel,
      },
    });
  }

  return results;
};

const normalizeCompileInstruction = (
  instruction: CompiledInstruction | MessageCompiledInstruction
): MessageCompiledInstruction => {
  if ("accounts" in instruction) {
    return {
      accountKeyIndexes: instruction.accounts,
      data: decode(instruction.data),
      programIdIndex: instruction.programIdIndex,
    };
  } else {
    return instruction;
  }
};
