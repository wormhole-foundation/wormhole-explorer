import { SuiEvent, SuiTransactionBlock } from "@mysten/sui.js/client";

// from https://github.com/MystenLabs/sui/blob/9336315169d3a2912260bbe298995af96a239953/sdk/typescript/src/client/types/generated.ts

export interface SuiTransactionBlockReceipt {
  checkpoint: string;
  digest: string;
  errors?: string[];
  events: SuiEvent[];
  timestampMs: string;
  transaction: SuiTransactionBlock;
}
