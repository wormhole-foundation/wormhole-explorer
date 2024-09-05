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
): Protocol => {
  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === chain) {
      const foundProtocolByAddress = contract.protocols.find((protocol) =>
        protocol.addresses.some((addr) => addr.toLowerCase() === address.toLowerCase())
      );

      if (!foundProtocolByAddress) {
        // Find the protocol that contains the method with the given comparativeMethod
        const foundProtocolByMethod = contract.protocols.find((protocol) =>
          protocol.methods.some((method) => method.methodId === comparativeMethod)
        );

        // Extract the method and type, providing default values if not found
        const method =
          foundProtocolByMethod?.methods.find((method) => method.methodId === comparativeMethod)
            ?.method ?? "unknown";
        const type = foundProtocolByMethod?.type ?? "unknown";

        return { method, type };
      }

      const foundMethod = foundProtocolByAddress?.methods.find(
        (method) => method.methodId === String(comparativeMethod)
      );

      if (foundMethod && foundProtocolByAddress) {
        return {
          method: foundMethod.method,
          type: foundProtocolByAddress.type,
        };
      }
    }
  }
  logger.warn(
    `[${chain}] Protocol not found, [hash: ${hash}][address: ${address}][method: ${comparativeMethod}]`
  );

  return {
    method: "unknown",
    type: "unknown",
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
