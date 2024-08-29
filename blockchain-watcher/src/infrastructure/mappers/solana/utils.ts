import { CompiledInstruction, MessageCompiledInstruction } from "@solana/web3.js";
import bs58 from "bs58";

export const normalizeCompileInstruction = (
  instruction: CompiledInstruction | MessageCompiledInstruction
): MessageCompiledInstruction => {
  if ("accounts" in instruction) {
    return {
      accountKeyIndexes: instruction.accounts,
      data: bs58.decode(instruction.data),
      programIdIndex: instruction.programIdIndex,
    };
  } else {
    return instruction;
  }
};
