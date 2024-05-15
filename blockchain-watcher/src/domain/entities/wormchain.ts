export type WormchainBlockLogs = {
  blockHeight: bigint;
  timestamp: number;
  chainId: number;
  transactions?: WormchainTransaction[];
};

export type WormchainTransaction = {
  height: string;
  hash: string;
  tx: Buffer;
  attributes: {
    key: string;
    value: string;
    index: boolean;
  }[];
};

export type WormchainTransactionByAttributes = {
  blockTimestamp: number;
  coreContract: string;
  targetChain: number;
  srcChannel: string;
  dstChannel: string;
  timestamp: string;
  receiver: string;
  sequence: number;
  sender: string;
  hash: string;
  tx: Buffer;
};

export type CosmosTransaction = {
  height: string;
  hash: string;
  tx: Buffer;
  attributes: {
    key: string;
    value: string;
    index: boolean;
  }[];
};

export type CosmosRedeem = {
  blockTimestamp: number;
  timestamp: string;
  chainId: number;
  height: string;
  hash: string;
  data: string;
  tx: Buffer;
  events: {
    type: string;
    attributes: [
      {
        key: string;
        value: string;
        index: boolean;
      }
    ];
  }[];
};
