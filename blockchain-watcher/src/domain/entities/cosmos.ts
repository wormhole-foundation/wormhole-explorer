export type CosmosTransaction = {
  blockTimestamp?: number;
  timestamp?: number;
  chainId?: number;
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
