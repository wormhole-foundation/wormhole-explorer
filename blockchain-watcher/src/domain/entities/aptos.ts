import { Types } from "aptos";

export type AptosEvent = Omit<Types.Event, "data"> & {
  version: string;
  data: {
    consistency_level: number;
    nonce: string;
    payload: string;
    sender: string;
    sequence: string;
    timestamp: string;
  };
};

export type AptosTransaction = {
  consistencyLevel: number;
  emitterChain?: number;
  blockHeight: bigint;
  timestamp: number;
  blockTime: number;
  sequence: bigint;
  version: string;
  payload: string;
  address: string;
  sender: string;
  status?: boolean;
  events: any;
  nonce: number;
  hash: string;
  type?: string;
};
