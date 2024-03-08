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

export enum TxStatus {
  Confirmed = "success",
  Failed = "failed",
}
