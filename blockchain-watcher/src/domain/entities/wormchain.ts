export type WormchainBlockLogs = {
  blockHeight: bigint;
  timestamp: number;
  chainId: number;
  transactions?: CosmosTransaction[];
};

export type CosmosTransaction = {
  hash: string;
  height: string;
  attributes: {
    key: string;
    value: string;
    index: boolean;
  }[];
};

// TODO: Watch this
export type TransactionAttributes = {
  coreContract: string | undefined;
  srcChannel: string | undefined;
  dstChannel: string | undefined;
  timestamp: string | undefined;
  receiver: string | undefined;
  sequence: number | undefined;
  chainId: number | undefined;
  sender: string | undefined;
  hash: string | undefined;
};

export type CosmosRedeem = {
  coreContract: string | undefined;
  srcChannel: string | undefined;
  dstChannel: string | undefined;
  timestamp: string | undefined;
  receiver: string | undefined;
  sequence: number | undefined;
  chainId: number | undefined;
  height: string | undefined;
  sender: string | undefined;
  hash: string | undefined;
};
