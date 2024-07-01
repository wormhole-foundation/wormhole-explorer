import {
  InstrumentedEthersProvider,
  InstrumentedConnection,
  WeightedProvidersPool,
  InstrumentedSuiClient,
} from "@xlabs/rpc-pool";

export class HealthyProvidersPool<
  T extends InstrumentedEthersProvider | InstrumentedConnection | InstrumentedSuiClient
> extends WeightedProvidersPool<T> {
  get(): T {
    const healthyProviders = this.getAllHealthy();
    if (healthyProviders && healthyProviders.length > 0) {
      return healthyProviders[0];
    }

    const unhealthyProviders = this.getAllUnhealthy();
    const randomProvider =
      unhealthyProviders[Math.floor(Math.random() * unhealthyProviders.length)];
    return randomProvider;
  }
}
