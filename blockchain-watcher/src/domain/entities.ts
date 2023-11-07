export type EvmBlock = {
  number: bigint;
  hash: string;
  timestamp: bigint; // epoch millis
};

export type EvmLog = {
  blockTime: bigint;
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
  blockTime: bigint;
  attributes: T;
};
