export type WormchainLog = {
  blockHeight: bigint;
  timestamp: number;
  transactions?: {
    hash: string;
    type: string;
    attributes: {
      index: boolean;
      value: string;
      key: string;
    }[];
  }[];
};
