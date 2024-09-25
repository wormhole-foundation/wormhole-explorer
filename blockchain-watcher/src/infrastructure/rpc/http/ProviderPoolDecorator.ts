import { InstrumentedConnectionWrapper } from "./InstrumentedConnectionWrapper";
import { InstrumentedSuiClientWrapper } from "./InstrumentedSuiClientWrapper";
import { InstrumentedHttpProvider } from "./InstrumentedHttpProvider";
import { HealthyProvidersPool } from "./HealthyProvidersPool";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
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
  setProviders(
    chain: string,
    providers:
      | InstrumentedHttpProvider[]
      | InstrumentedConnectionWrapper[]
      | InstrumentedSuiClientWrapper[],
    providersHealthCheck: ProviderHealthCheck[],
    blockHeightCursor: bigint | undefined
  ): void;
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
      return providerPoolSupplier(
        rpcs,
        createProvider,
        type,
        logger
      ) as unknown as ProviderPoolDecorator<T>;
  }
}
