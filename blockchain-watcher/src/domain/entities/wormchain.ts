export type WormchainLog = {
  blockHeight: bigint;
  timestamp: number;
  transactions?: {
    hash: string;
    type: string;
    attributes: {
      key: string;
      value: string;
      index: boolean;
    }[];
  }[];
};
