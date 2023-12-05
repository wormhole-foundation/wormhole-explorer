import { decode } from "bs58";
import { Connection, Commitment } from "@solana/web3.js";
import { solana, LogFoundEvent, TransferRedeemed } from "../../domain/entities";
import { CompiledInstruction, MessageCompiledInstruction } from "../../domain/entities/solana";
import { configuration } from "../config";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";

enum Instruction {
  CompleteNativeTransferWithRelay = 0x02,
  CompleteWrappedTransferWithRelay = 0x03,
}

const connection = new Connection(configuration.chains.solana.rpcs[0]);

export const solanaTransferRedeemedMapper = async (
  tx: solana.Transaction,
  { programId, commitment }: { programId: string; commitment?: Commitment }
): Promise<LogFoundEvent<TransferRedeemed>[]> => {
  if (!tx || !tx.blockTime) {
    throw new Error(`Block time is missing for tx in slot ${tx?.slot} @ time ${tx?.blockTime}`);
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

  const results: LogFoundEvent<TransferRedeemed>[] = [];
  for (const instruction of whInstructions) {
    const instructionId = instruction.data;
    if (
      instructionId[0] !== Instruction.CompleteNativeTransferWithRelay &&
      instructionId[0] !== Instruction.CompleteWrappedTransferWithRelay
    ) {
      continue;
    }
    const accountAddress = accountKeys[instruction.accountKeyIndexes[2]];

    const { message } = await getPostedMessage(connection, accountAddress, commitment);
    const { sequence, emitterAddress, emitterChain } = message || {};

    results.push({
      name: "transfer-redeemed",
      address: programId,
      chainId: 1,
      txHash: tx.transaction.signatures[0],
      blockHeight: BigInt(tx.slot.toString()),
      blockTime: tx.blockTime,
      attributes: {
        emitterChainId: emitterChain,
        emitterAddress: emitterAddress.toString("hex"),
        sequence: Number(sequence),
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
