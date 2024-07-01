import { InstrumentedRpc, RpcConfig, providerPoolSupplier } from "@xlabs/rpc-pool";
import { Logger } from "winston";

export interface ProviderPoolDecorator<T extends InstrumentedRpc> {
  getAllUnhealthy(): T[];
  getAllHealthy(): T[];
  getProvider(): T;
  get(): T;
}

export function providerPoolSupplierDecorator<T extends InstrumentedRpc>(
  rpcs: RpcConfig[],
  createProvider: (rpcCfg: RpcConfig) => T,
  type?: string,
  logger?: Logger
): ProviderPoolDecorator<T> {
  const result = providerPoolSupplier(
    rpcs,
    createProvider,
    type,
    logger
  ) as ProviderPoolDecorator<T>;

  result.getProvider = function () {
    const healthyProviders = this.getAllHealthy();
    if (healthyProviders && healthyProviders.length > 0) {
      return healthyProviders[0];
    }

    const unhealthyProviders = this.getAllUnhealthy();
    const randomProvider =
      unhealthyProviders[Math.floor(Math.random() * unhealthyProviders.length)];
    return randomProvider;
  };

  return result;
}
