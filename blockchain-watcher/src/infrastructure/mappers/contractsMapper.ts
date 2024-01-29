import * as configData from "./contractsMapperConfig.json";

export const contractsMapperConfig: ContractsMapperConfig = configData as ContractsMapperConfig;

export type Protocol = {
  method: string;
  type: string;
};

export interface ContractsMapperConfig {
  contracts: {
    chain: string;
    protocols: {
      address: string[];
      type: string;
      methods: { methodId: string; method: string }[];
    }[];
  }[];
}
