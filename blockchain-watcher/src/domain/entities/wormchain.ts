export type WormchainBlockLogs = {
  transactions?: CosmosTransaction[];
  blockHeight: bigint;
  timestamp: number;
};

export type IbcTransaction = {
  gatewayContract: string;
  blockTimestamp: number;
  targetChain: number; // (osmosis, kujira, injective, evmos etc)
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
  height: bigint;
  hash: string;
  tx: Buffer;
  attributes: {
    index: boolean;
    value: string;
    key: string;
  }[];
};

export type CosmosRedeem = {
  blockTimestamp?: number;
  timestamp?: string;
  chainId: number;
  height: bigint;
  chain?: string;
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
