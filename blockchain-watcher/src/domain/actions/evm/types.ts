export interface HandleEvmConfig {
  environment: string;
  metricName: string;
  commitment: string;
  chainId: number;
  chain: string;
  abis: Abis;
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

export type Abis = {
  topic: string;
  abi: string;
}[];
