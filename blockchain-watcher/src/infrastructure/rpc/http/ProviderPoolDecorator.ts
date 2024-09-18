import { HealthyProvidersPool } from "./HealthyProvidersPool";
import { Logger } from "winston";
import {
  InstrumentedEthersProvider,
  InstrumentedConnection,
  providerPoolSupplier,
  InstrumentedRpc,
  ProviderPool,
  RpcConfig,
} from "@xlabs/rpc-pool";

export interface ProviderPoolDecorator<T extends InstrumentedRpc> extends ProviderPool<T> {
  getProviders(): T[];
  setProviders(): void;
}

export function extendedProviderPoolSupplier<T extends InstrumentedRpc>(
  rpcs: RpcConfig[],
  createProvider: (rpcCfg: RpcConfig) => T,
  type?: string,
  logger?: Logger
): ProviderPoolDecorator<T> {
  switch (type) {
    case "healthy":
      return HealthyProvidersPool.fromConfigs(
        rpcs,
        createProvider as unknown as (
          rpc: RpcConfig
        ) => InstrumentedEthersProvider | InstrumentedConnection,
        logger
      ) as unknown as ProviderPoolDecorator<T>;
    default:
      return HealthyProvidersPool.fromConfigs(
        rpcs,
        createProvider as unknown as (
          rpc: RpcConfig
        ) => InstrumentedEthersProvider | InstrumentedConnection,
        logger
      ) as unknown as ProviderPoolDecorator<T>;
  }
}
