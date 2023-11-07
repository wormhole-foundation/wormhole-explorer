export abstract class Watcher<Output, Cfg> {
  private name: string;
  private environment: string;
  private chain: string;

  constructor(name: string, environment: string, chain: string) {
    this.name = name;
    this.environment = environment;
    this.chain = chain;
  }

  getConfiguration(): Cfg {
    throw new Error("Method not implemented.");
  }

  abstract watch(handlers: Handler<Output, any>[]): Promise<void>;
}

export interface Handler<Input, Output> {
  handle(input: Input[]): Promise<HandlerResult<Output>>;
}

export class HandlerResult<Output> {
  constructor(public readonly output: Output) {}
}

export type EvmBlock = {
  number: bigint;
  hash: string;
  timestamp: bigint; // epoch millis
};

export type EvmLog = {
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
