export type Transaction = {
  chainId: number;
  chain: string;
  slot: number;
  transaction: {
    message: Message;
    signatures: string[];
  };
  meta?: {
    innerInstructions?: CompiledInnerInstruction[] | null;
    fee: number;
    err?: {} | string | null;
  };
  blockTime?: number | null;
};

type CompiledInnerInstruction = {
  index: number;
  instructions: CompiledInstruction[];
};

export type CompiledInstruction = {
  programIdIndex: number;
  accounts: number[];
  data: string;
};

export type Message = {
  accountKeys: string[];
  instructions: CompiledInstruction[];
  compiledInstructions: MessageCompiledInstruction[];
};

export type Block = {
  /** Blockhash of this block */
  blockhash: string;
  transactions: {
    transaction: {
      message: Message;
      signatures: string[];
    };
  }[];
  blockTime: number | null;
  blockHeight: number | null;
};

export type MessageCompiledInstruction = {
  /** Index into the transaction keys array indicating the program account that executes this instruction */
  programIdIndex: number;
  /** Ordered indices into the transaction keys array indicating which accounts to pass to the program */
  accountKeyIndexes: number[];
  /** The program input data */
  data: Uint8Array;
};

export type ConfirmedSignatureInfo = {
  signature: string;
  err?: any | null;
  blockTime?: number | null;
};

export enum ErrorType {
  SkippedSlot,
  NoBlockOrBlockTime,
}
