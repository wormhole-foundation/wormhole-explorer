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

export type LogFoundEvent<T> = {
  name: string;
  chainId: number;
  txHash: string;
  blockHeight: bigint;
  blockTime: number;
  attributes: T;
};

export type LogMessagePublished = {
  sequence: number;
  sender: string;
  nonce: number;
  payload: string;
  consistencyLevel: number;
};
