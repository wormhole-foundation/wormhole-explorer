import { InstrumentedHttpProvider } from "./InstrumentedHttpProvider";
import { ProvidersHeight } from "./ProviderPoolDecorator";
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

const THRESHOLD = 100n;

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
    providersHeight: ProvidersHeight[],
    blockHeightCursor: bigint | undefined
  ): void {
    if (!providersHeight || providersHeight.length === 0) {
      this.providers = this.get() as unknown as T[];
      return;
    }

    const sortedProviders = this.sortResponsesByHeight(providersHeight);
    const averageHeight = this.calculateAverageHeight(sortedProviders);
    const filteredProviders = this.filterOutliers(
      sortedProviders,
      averageHeight,
      blockHeightCursor
    );
    const healthyProviders = this.removeInvalidProviders(providers, filteredProviders);

    this.providers =
      healthyProviders.length > 0
        ? (healthyProviders as unknown as T[])
        : (this.get() as unknown as T[]);
  }

  private sortResponsesByHeight(providersHeight: ProvidersHeight[]): ProvidersHeight[] {
    return providersHeight.sort((a, b) => Number(b.height) - Number(a.height));
  }

  private calculateAverageHeight(providers: ProvidersHeight[]): bigint {
    const totalHeight = providers.reduce((sum, provider) => sum + provider.height, BigInt(0));
    return totalHeight / BigInt(providers.length);
  }

  private filterOutliers(
    providers: ProvidersHeight[],
    averageHeight: bigint,
    blockHeightCursor: bigint | undefined
  ): ProvidersHeight[] {
    return providers.filter((provider) => {
      // Check if the provider is behind the cursor
      if (blockHeightCursor && provider.height < blockHeightCursor) {
        logger.warn(
          `Provider ${provider.url} is not healthy: behind the cursor [${provider.height} < ${blockHeightCursor}]`
        );
        return false;
      }

      // Calculate the deviation from the average height
      const deviation = Math.abs(Number(provider.height) - Number(averageHeight));

      // Check if the provider's height deviates too much from the average height
      if (deviation > THRESHOLD) {
        logger.warn(
          `Provider ${provider.url} is not healthy: deviation from average height is too high [${deviation} > ${THRESHOLD}]`
        );
        return false;
      }

      // If the provider passes both checks, it is considered healthy
      return true;
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
