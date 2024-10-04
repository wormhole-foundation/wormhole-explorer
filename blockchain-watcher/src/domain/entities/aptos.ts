import { Types } from "aptos";

export type AptosEvent = Omit<Types.Event, "data"> & {
  hash?: string;
  events?: any;
  success?: boolean;
  timestamp?: string;
  version?: string;
  payload?: any;
  data: {
    consistency_level: number;
    nonce: string;
    payload: any;
    sender: string;
    sequence: string;
    timestamp: string;
  };
};

export type AptosTransaction = {
  blockHeight: bigint;
  timestamp?: number;
  blockTime?: number;
  version: string;
  payload?: any;
  status?: boolean;
  events: any;
  nonce?: number;
  hash: string;
  type?: string;
  sequence_number?: string;
};

export type AptosTransactionByVersion = {
  sequence_number?: string;
  timestamp?: string;
  success?: boolean;
  version?: string;
  payload?: any;
  events?: any[];
  sender?: string;
  hash?: string;
};

export type AptosBlockByVersion = {
  block_height?: string;
};

export type LedgerInfo = {
  ledger_version: string;
};
