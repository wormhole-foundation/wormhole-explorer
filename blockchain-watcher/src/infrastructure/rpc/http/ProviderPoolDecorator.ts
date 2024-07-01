import { InstrumentedRpc, ProviderPool, RpcConfig, providerPoolSupplier } from "@xlabs/rpc-pool";
import { Logger } from "winston";

export function providerPoolSupplierDecorator<T extends InstrumentedRpc>(
  rpcs: RpcConfig[],
  createProvider: (rpcCfg: RpcConfig) => T,
  type?: string,
  logger?: Logger
): ProviderPool<T> {
  const result = providerPoolSupplier(rpcs, createProvider, type, logger) as ProviderPool<T>;

  result.get = function () {
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
