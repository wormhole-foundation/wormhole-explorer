import {
  CompiledInstruction,
  Message,
  MessageCompiledInstruction,
  MessageV0,
} from '@solana/web3.js';
import { decode } from 'bs58';

export const isLegacyMessage = (message: Message | MessageV0): message is Message => {
  return message.version === 'legacy';
};

export const normalizeCompileInstruction = (
  instruction: CompiledInstruction | MessageCompiledInstruction
): MessageCompiledInstruction => {
  if ('accounts' in instruction) {
    return {
      accountKeyIndexes: instruction.accounts,
      data: decode(instruction.data),
      programIdIndex: instruction.programIdIndex,
    };
  } else {
    return instruction;
  }
};
