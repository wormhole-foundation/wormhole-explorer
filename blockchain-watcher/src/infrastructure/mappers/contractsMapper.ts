import * as configData from "./contractsMapperConfig.json";
import winston from "../log";

const UNKNOWN = "unknown";

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
): Protocol => {
  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === chain) {
      // Find the protocol by address
      let protocol = contract.protocols.find((protocol) =>
        protocol.addresses.some((addr) => addr.toLowerCase() === address.toLowerCase())
      );

      if (!protocol) {
        // If not found by address, try to find by method
        const protocolsByMethod = contract.protocols.filter((protocol) =>
          protocol.methods.some((method) => method.methodId === comparativeMethod)
        );

        // If not found by method or found more than one, return unknown
        if (!protocolsByMethod || protocolsByMethod.length > 1) {
          return {
            method: UNKNOWN,
            type: UNKNOWN,
          };
        }

        protocol = protocolsByMethod[0];
      }

      // Find the method in the identified protocol
      const method = protocol.methods.find(
        (method) => method.methodId === String(comparativeMethod)
      );

      return {
        method: method?.method ?? UNKNOWN,
        type: protocol.type ?? UNKNOWN,
      };
    }
  }

  logger.warn(
    `[${chain}] Protocol not found, [hash: ${hash}][address: ${address}][method: ${comparativeMethod}]`
  );

  return {
    method: UNKNOWN,
    type: UNKNOWN,
  };
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
