export type EvmBlock = {
  number: bigint;
  hash: string;
  timestamp: number; // epoch seconds
};

export type EvmLog = {
  blockTime?: number; // epoch seconds
  blockNumber: bigint;
  blockHash: string;
  address: string;
  removed: boolean;
  data: string;
  transactionHash: string;
  transactionIndex: string;
  topics: string[];
  logIndex: number;
  chainId?: number;
};

export type EvmTag = "finalized" | "latest" | "safe";

export type EvmTopicFilter = {
  addresses: string[];
  topics: string[];
};

export type EvmLogFilter = {
  fromBlock: bigint | EvmTag;
  toBlock: bigint | EvmTag;
  addresses: string[];
  topics: string[];
};
