export type Transaction = {
  slot: number;
  transaction: {
    message: Message;
    signatures: string[];
  };
  meta?: {
    innerInstructions?: CompiledInnerInstruction[] | null;
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

export class Failure extends Error {
  readonly code?: number | unknown;
  readonly type?: ErrorType;
  constructor(code: number | unknown, message: string, type?: ErrorType) {
    super(message);
    this.code = code;

    if (this.code === -32007 || this.code === -32009) {
      this.type = ErrorType.SkippedSlot;
    }

    if (type) {
      this.type = type;
    }
  }

  public skippedSlot(): boolean {
    return this.type === ErrorType.SkippedSlot;
  }

  public noBlockOrBlockTime(): boolean {
    return this.type === ErrorType.NoBlockOrBlockTime;
  }
}
