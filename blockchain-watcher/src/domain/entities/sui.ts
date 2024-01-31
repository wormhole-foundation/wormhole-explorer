import { ExecutionStatus, SuiEvent, SuiTransactionBlock } from "@mysten/sui.js/client";

export interface SuiTransactionBlockReceipt {
  checkpoint: string;
  digest: string;
  effects?: {
    status?: ExecutionStatus;
  };
  events: SuiEvent[];
  timestampMs: string;
  transaction: SuiTransactionBlock;
}
