export type WormchainBlockLogs = {
  blockHeight: bigint;
  timestamp: number;
  chainId: number;
  transactions?: CosmosTransaction[];
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
  vaaEmitterAddress: string;
  vaaEmitterChain: number;
  blockTimestamp: number;
  vaaSequence: bigint;
  timestamp: string;
  chainId: number;
  height: string;
  hash: string;
  data: string;
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

export type WormchainTransaction = {
  vaaEmitterAddress: string;
  vaaEmitterChain: number;
  blockTimestamp: number;
  coreContract: string;
  targetChain: number;
  vaaSequence: bigint;
  srcChannel: string;
  dstChannel: string;
  timestamp: string;
  receiver: string;
  sequence: number;
  sender: string;
  hash: string;
};
