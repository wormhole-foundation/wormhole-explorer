import { InstrumentedHttpProvider } from "./InstrumentedHttpProvider";
import { ProvidersHeight } from "./ProviderPoolDecorator";
import { Logger } from "winston";
import {
  InstrumentedEthersProvider,
  InstrumentedConnection,
  WeightedProvidersPool,
  InstrumentedSuiClient,
  ProviderPoolStrategy,
  RpcConfig,
} from "@xlabs/rpc-pool";

type Weighted<T> = { provider: T; weight: number };

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

  static fromConfigs<
    T extends InstrumentedEthersProvider | InstrumentedConnection | InstrumentedSuiClient
  >(
    rpcs: RpcConfig[],
    createProvider: (rpcCfg: RpcConfig) => T,
    logger?: Logger
  ): ProviderPoolStrategy<T> {
    const providers: Weighted<T>[] = [];
    for (const rpcCfg of rpcs) {
      providers.push({ provider: createProvider(rpcCfg), weight: rpcCfg.weight ?? 1 });
    }
    return new HealthyProvidersPool(providers, logger);
  }

  getProviders(): T[] {
    return this.providers;
  }

  setProviders(providers: InstrumentedHttpProvider[], providersHeight: ProvidersHeight[]): void {
    const sortedProviders = this.sortResponsesByHeight(providersHeight);
    const averageHeight = this.calculateAverageHeight(sortedProviders);
    const filteredProviders = this.filterOutliers(sortedProviders, averageHeight, 5);
    const healthyProviders = this.removeInvalidProviders(providers, filteredProviders);

    this.providers =
      healthyProviders.length > 0
        ? (healthyProviders as unknown as T[])
        : (this.get() as unknown as T[]);
  }

  private sortResponsesByHeight(providersHeight: ProvidersHeight[]): ProvidersHeight[] {
    return providersHeight.sort((a, b) => Number(b.height) - Number(a.height));
  }

  private calculateAverageHeight(providersHeight: ProvidersHeight[]): number {
    if (providersHeight.length === 0) return 0;

    const totalHeight = providersHeight.reduce((sum, provider) => sum + Number(provider.height), 0);
    return totalHeight / providersHeight.length;
  }

  private filterOutliers(
    providersHeight: ProvidersHeight[],
    averageHeight: number,
    threshold: number
  ): ProvidersHeight[] {
    return providersHeight.filter((provider) => {
      const deviation =
        Number(provider.height) > averageHeight
          ? Number(provider.height) - averageHeight
          : averageHeight - Number(provider.height);
      return deviation <= threshold;
    });
  }

  private removeInvalidProviders(
    providers: InstrumentedHttpProvider[],
    filteredProviders: ProvidersHeight[]
  ): InstrumentedHttpProvider[] {
    const filteredUrls = new Set(filteredProviders.map((provider) => provider.url));
    return providers.filter((provider) => filteredUrls.has(provider.getUrl()));
  }
}
