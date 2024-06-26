export type SeiRedeem = {
  timestamp?: number;
  chainId: number;
  height: bigint;
  hash: string;
  data: string;
  tx: Buffer;
  events: {
    type: string;
    attributes: {
      index: boolean;
      value: string;
      key: string;
    }[];
  }[];
};

export type WormchainRedeem = {
  timestamp?: number;
  chainId: number;
  height: bigint;
  hash: string;
  data: string;
  tx: Buffer;
  events: {
    type: string;
    attributes: {
      index: boolean;
      value: string;
      key: string;
    }[];
  }[];
};
