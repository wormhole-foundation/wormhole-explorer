import { InstrumentedHttpProvider } from "./InstrumentedHttpProvider";
import { ProviderHealthCheck } from "../../../domain/actions/poolRpcs/PoolRpcs";
import { Logger } from "winston";
import winston from "../../log";
import {
  InstrumentedEthersProvider,
  InstrumentedConnection,
  WeightedProvidersPool,
  InstrumentedSuiClient,
  ProviderPoolStrategy,
  RpcConfig,
} from "@xlabs/rpc-pool";

let logger: winston.Logger;
logger = winston.child({ module: "HealthyProvidersPool" });

type Weighted<T> = { provider: T; weight: number };

const THRESHOLD = 5_000n;

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

  setProviders(
    providers: InstrumentedHttpProvider[],
    providerHealthCheck: ProviderHealthCheck[],
    blockHeightCursor: bigint | undefined
  ): void {
    const auxProvider = providers;

    // If there are no providers or cursor, we dont need to do anything
    if (!blockHeightCursor || !providerHealthCheck || providerHealthCheck.length === 0) {
      return;
    }

    const filteredProviders = this.filterProviders(providerHealthCheck, blockHeightCursor);
    if (filteredProviders.length === 0) {
      return;
    }

    const sortedProviders = this.sortResponsesByHeight(filteredProviders);
    const averageProviders = this.calculateAverageHeight(filteredProviders);
    const healthyProviders = this.removeInvalidProviders(auxProvider, sortedProviders);

    this.providers = (healthyProviders as unknown as T[]) ?? providers;
  }

  private filterProviders(providers: ProviderHealthCheck[], blockHeightCursor: bigint) {
    // Filter out providers that are behind the cursor
    let providersToFilter = providers.filter((provider) => provider.height! >= blockHeightCursor);
    const heights = providersToFilter.map((item) => parseFloat(String(item.height)));

    // Determine the maximum height and the next maximum height
    const maxHeight = Math.max(...heights);
    const nextMaxHeight = Math.max(...heights.filter((h) => h < maxHeight)); // TODO: This is not working as expected

    // Filter out the maximum height if it's significantly ahead
    if (!nextMaxHeight && maxHeight - nextMaxHeight > THRESHOLD) {
      providersToFilter = providersToFilter.filter(
        (item) => parseFloat(String(item.height)) < nextMaxHeight + 1
      );
    }
    return providersToFilter;
  }

  private sortResponsesByHeight(providerHealthCheck: ProviderHealthCheck[]): ProviderHealthCheck[] {
    return providerHealthCheck.sort((a, b) => Number(b.height) - Number(a.height));
  }

  private calculateAverageHeight(filteredProviders: ProviderHealthCheck[]): bigint {
    const totalHeight = filteredProviders.reduce(
      (sum, provider) => sum + provider.height!,
      BigInt(0)
    );
    const a = totalHeight / BigInt(filteredProviders.length);
    return a;
  }

  private removeInvalidProviders(
    auxProvider: InstrumentedHttpProvider[],
    filteredProviders: ProviderHealthCheck[]
  ): InstrumentedHttpProvider[] {
    const filteredUrls = new Set(filteredProviders.map((provider) => provider.url));
    return auxProvider.filter((provider) => filteredUrls.has(provider.getUrl()));
  }
}
