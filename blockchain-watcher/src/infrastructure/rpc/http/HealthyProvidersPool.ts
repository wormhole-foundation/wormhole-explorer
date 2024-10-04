import { InstrumentedHttpProvider } from "./InstrumentedHttpProvider";
import { ProviderHealthCheck } from "../../../domain/repositories";
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
    chain: string,
    providers: InstrumentedHttpProvider[],
    providersHealthCheck: ProviderHealthCheck[],
    cursor: bigint | undefined
  ): void {
    if (providers?.length === 0 || providersHealthCheck?.length === 0) {
      return;
    }
    const providersHealthy = providersHealthCheck.filter((provider) => provider.isHealthy);
    if (providersHealthy.length === 0) {
      return;
    }

    const filter = this.filter(providersHealthy, cursor);
    const sort = this.sort(filter);
    const healthy = this.setOffline(providers, sort);

    logger.info(
      `[${chain}] Healthy providers: ${healthy
        .map((provider) => (provider.isHealthy() ? provider.getUrl() : undefined))
        ?.join(", ")}`
    );
    this.providers =
      healthy.length > 0 ? (healthy as unknown as T[]) : (providers as unknown as T[]);
  }

  private filter(
    providersHealthCheck: ProviderHealthCheck[],
    cursor: bigint | undefined
  ): ProviderHealthCheck[] {
    // Filter out providers that are behind the cursor
    if (cursor) {
      providersHealthCheck = providersHealthCheck.filter(
        (provider) => provider.height && provider.height >= cursor
      );
    }

    const providerWithHeight = providersHealthCheck.filter((provider) => provider.height);
    if (providerWithHeight?.length > 0) {
      const heights = providerWithHeight.map((item) => parseFloat(String(item.height)));

      // Determine the maximum height and the next maximum height
      const maxHeight = Math.max(...heights);
      const nextMaxHeight = Math.max(...heights.filter((h) => h < maxHeight));

      // Filter out the maximum height if it's significantly ahead
      if (!nextMaxHeight && maxHeight - nextMaxHeight > THRESHOLD) {
        providersHealthCheck = providerWithHeight.filter(
          (item) => parseFloat(String(item.height)) < maxHeight
        );
      }
    }
    return providersHealthCheck;
  }

  private sort(providersHealthCheck: ProviderHealthCheck[]): ProviderHealthCheck[] {
    return providersHealthCheck.sort((a, b) => {
      const heightDifference = Number(b.height) - Number(a.height);
      if (heightDifference !== 0) {
        return heightDifference;
      }
      // Handle cases where latency might be undefined
      const latencyA = a.latency ?? Infinity;
      const latencyB = b.latency ?? Infinity;
      return latencyA - latencyB;
    });
  }

  // Only set up offline the providers because we can lose all the providers
  // and the pool will be empty
  private setOffline(
    providers: InstrumentedHttpProvider[],
    providersHealthCheck: ProviderHealthCheck[]
  ): InstrumentedHttpProvider[] {
    const filteredUrlIndexMap = new Map<string, number>();
    providersHealthCheck.forEach((provider, index) => {
      filteredUrlIndexMap.set(provider.url, index);
    });

    return providers
      .map((provider) => {
        // If the provider is not in the filtered list, set it offline
        if (!filteredUrlIndexMap.has(provider.getUrl())) {
          provider.setProviderOffline();
        }
        return provider;
      })
      .sort((a, b) => {
        // Sort the providers by the order of the filtered list
        const indexA = filteredUrlIndexMap.get(a.getUrl())!;
        const indexB = filteredUrlIndexMap.get(b.getUrl())!;
        return indexA - indexB;
      });
  }
}
