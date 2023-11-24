export enum ErrorType {
  SkippedSlot,
  NoBlockOrBlockTime,
}

export class SolanaFailure extends Error {
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
