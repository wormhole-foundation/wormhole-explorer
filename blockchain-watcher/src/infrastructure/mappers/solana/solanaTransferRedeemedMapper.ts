import { solana, TransactionFoundEvent, SolanaTransactionFound } from "../../../domain/entities";
import { CompiledInstruction, MessageCompiledInstruction } from "../../../domain/entities/solana";
import { methodNameByInstructionMapper } from "./methodNameByInstructionMapper";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { configuration } from "../../config";
import { decode } from "bs58";

enum Instruction {
  CompleteNativeTransfer = 0x02,
  CompleteWrappedTransfer = 0x03,
  CompleteNativeWithPayload = 0x09,
  CompleteWrappedWithPayload = 0x0a,
}

const TRANSACTION_STATUS_COMPLETED = "completed";
const TRANSACTION_STATUS_FAILED = "failed";

const connection = new Connection(configuration.chains.solana.rpcs[0]);

export const solanaTransferRedeemedMapper = async (
  tx: solana.Transaction,
  { programId, commitment }: { programId: string; commitment?: Commitment }
): Promise<TransactionFoundEvent<SolanaTransactionFound>[]> => {
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

  const results: TransactionFoundEvent<SolanaTransactionFound>[] = [];
  for (const instruction of whInstructions) {
    if (isNotACompleteTransferInstruction(instruction.data)) {
      continue;
    }

    const accountAddress = accountKeys[instruction.accountKeyIndexes[2]];
    const { message } = await getPostedMessage(connection, accountAddress, commitment);
    const { sequence, emitterAddress, emitterChain } = message || {};
    const methods = methodNameByInstructionMapper(instruction, programIdIndex);

    results.push({
      name: "transfer-redeemed",
      address: programId,
      chainId: 1,
      txHash: tx.transaction.signatures[0],
      blockHeight: BigInt(tx.slot.toString()),
      blockTime: tx.blockTime,
      attributes: {
        method: methods.method,
        status: mappedStatus(tx),
        emitterChainId: emitterChain,
        emitterAddress: emitterAddress.toString("hex"),
        sequence: Number(sequence),
      },
    });
  }

  return results;
};

const mappedStatus = (tx: solana.Transaction): string => {
  if (!tx.meta || tx.meta.err) TRANSACTION_STATUS_FAILED;
  return TRANSACTION_STATUS_COMPLETED;
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

/**
 * Checks if the instruction is not to complete a transfer.
 * @param instructionId - the instruction id
 * @returns true if the instruction is valid, false otherwise
 */
const isNotACompleteTransferInstruction = (instructionId: Uint8Array): boolean => {
  return (
    instructionId[0] !== Instruction.CompleteNativeTransfer &&
    instructionId[0] !== Instruction.CompleteWrappedTransfer &&
    instructionId[0] !== Instruction.CompleteNativeWithPayload &&
    instructionId[0] !== Instruction.CompleteWrappedWithPayload
  );
};
