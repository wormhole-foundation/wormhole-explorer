export type WormchainBlockLogs = {
  blockHeight: bigint;
  timestamp: number;
  chainId: number;
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
