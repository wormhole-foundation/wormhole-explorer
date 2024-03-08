import { EvmTopicFilter } from "../../entities";

export interface HandleEvmConfig {
  metricName: string;
  commitment: string;
  chainId: number;
  chain: string;
  abi: string;
  id: string;
}

export interface HandleEvmLogsConfig extends HandleEvmConfig {
  filter: EvmTopicFilter;
}
