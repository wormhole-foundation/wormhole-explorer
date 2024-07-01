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

export function extendedProviderPoolSupplier<T extends InstrumentedRpc>(
  rpcs: RpcConfig[],
  createProvider: (rpcCfg: RpcConfig) => T,
  type?: string,
  logger?: Logger
): ProviderPool<T> {
  switch (type) {
    case "healthy":
      return HealthyProvidersPool.fromConfigs(
        rpcs,
        createProvider as unknown as (
          rpc: RpcConfig
        ) => InstrumentedEthersProvider | InstrumentedConnection,
        logger
      ) as unknown as ProviderPool<T>;
    default:
      return providerPoolSupplier(rpcs, createProvider, type, logger);
  }
}
