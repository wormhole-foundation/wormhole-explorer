import * as configData from "./contractsMapperConfig.json";
import winston from "../log";

let logger: winston.Logger;
logger = winston.child({ module: "contractsMapperConfig" });

export const contractsMapperConfig: ContractsMapperConfig = configData as ContractsMapperConfig;

export type Protocol = {
  method: string;
  type: string;
};

export const findProtocol = (
  chain: string,
  address: string,
  comparativeMethod: string | number,
  hash: string
): Protocol | undefined => {
  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === chain) {
      const foundProtocol = contract.protocols.find((protocol) =>
        protocol.addresses.some((addr) => addr.toLowerCase() === address.toLowerCase())
      );
      const foundMethod = foundProtocol?.methods.find(
        (method) => method.methodId === String(comparativeMethod)
      );

      if (foundMethod && foundProtocol) {
        return {
          method: foundMethod.method,
          type: foundProtocol.type,
        };
      }
    }
  }

  logger.warn(
    `[${chain}] Protocol not found, [tx hash: ${hash}][address: ${address}][method: ${comparativeMethod}]`
  );
};

export interface ContractsMapperConfig {
  contracts: {
    chain: string;
    protocols: {
      addresses: string[];
      type: string;
      methods: { methodId: string; method: string }[];
    }[];
  }[];
}
