export interface HandleEvmConfig {
  environment: string;
  metricName: string;
  commitment: string;
  chainId: number;
  chain: string;
  abis: Abi[];
  id: string;
}

export interface HandleEvmLogsConfig extends HandleEvmConfig {
  filters: Filters;
}

export type Filters = {
  addresses: string[];
  strategy?: string;
  topics: string[];
}[];

export type Abi = {
  topic: string;
  abi: string;
};
