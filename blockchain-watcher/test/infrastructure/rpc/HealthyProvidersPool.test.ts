import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool(false); // No mock healthy providers pool

import { InstrumentedEthersProvider, RpcConfig } from "@xlabs/rpc-pool";
import { jest, describe, it, expect } from "@jest/globals";
import { ProviderPoolDecorator } from "../../../src/infrastructure/rpc/http/ProviderPoolDecorator";
import { HealthyProvidersPool } from "../../../src/infrastructure/rpc/http/HealthyProvidersPool";
import { Logger } from "winston";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";

describe("HealthyProvidersPool", () => {
  it("should be order by height", () => {
    // Given
    const rpcConfigs: RpcConfig[] = [
      { url: "https://rpc1/v1", weight: 1 },
      { url: "https://rpc2/v1", weight: 1 },
      { url: "https://rpc3/v1", weight: 1 },
    ];

    const providers = [
      new InstrumentedHttpProvider({ url: "https://rpc1/v1", chain: "ethereum" }),
      new InstrumentedHttpProvider({ url: "https://rpc2/v1", chain: "ethereum" }),
    ];

    const providersHealthCheck = [
      {
        isHealthy: true,
        latency: 0.05,
        height: 100n,
        url: "https://rpc1/v1",
      },
      {
        isHealthy: true,
        latency: 0.07,
        height: 110n,
        url: "https://rpc2/v1",
      },
    ];

    // Mock createProvider function
    const createProvider = jest.fn((rpcCfg: RpcConfig) => ({
      getUrl: () => rpcCfg.url,
    })) as unknown as (rpcCfg: RpcConfig) => InstrumentedEthersProvider;

    // Call fromConfigs
    const pool = HealthyProvidersPool.fromConfigs(
      rpcConfigs,
      createProvider,
      jest.fn() as unknown as Logger
    ) as unknown as ProviderPoolDecorator<InstrumentedEthersProvider>;

    // When
    pool.setProviders("ethereum", providers, providersHealthCheck, 90n);
    const result = pool.getProviders() as unknown as InstrumentedHttpProvider[];

    // Then
    expect(pool).toBeInstanceOf(HealthyProvidersPool);
    expect(JSON.stringify(result[0])).toBe(
      '{"initialDelay":1000,"maxDelay":60000,"retries":0,"timeout":5000,"url":"https://rpc2/v1","chain":"ethereum","health":{"lastRequestDurations":[0.1232]},"logger":{}}'
    );
    expect(JSON.stringify(result[1])).toBe(
      '{"initialDelay":1000,"maxDelay":60000,"retries":0,"timeout":5000,"url":"https://rpc1/v1","chain":"ethereum","health":{"lastRequestDurations":[0.1232]},"logger":{}}'
    );
  });

  it("should be order by latency", () => {
    // Given
    const rpcConfigs: RpcConfig[] = [
      { url: "https://rpc1/v1", weight: 1 },
      { url: "https://rpc2/v1", weight: 1 },
      { url: "https://rpc3/v1", weight: 1 },
    ];

    const providers = [
      new InstrumentedHttpProvider({ url: "https://rpc1/v1", chain: "ethereum" }),
      new InstrumentedHttpProvider({ url: "https://rpc2/v1", chain: "ethereum" }),
      new InstrumentedHttpProvider({ url: "https://rpc3/v1", chain: "ethereum" }),
    ];

    const providersHealthCheck = [
      {
        isHealthy: true,
        latency: 0.05,
        height: 100n,
        url: "https://rpc1/v1",
      },
      {
        isHealthy: true,
        latency: 0.07,
        height: 100n,
        url: "https://rpc2/v1",
      },
      {
        isHealthy: true,
        latency: 0.06,
        height: 100n,
        url: "https://rpc3/v1",
      },
    ];

    // Mock createProvider function
    const createProvider = jest.fn((rpcCfg: RpcConfig) => ({
      getUrl: () => rpcCfg.url,
    })) as unknown as (rpcCfg: RpcConfig) => InstrumentedEthersProvider;

    // Call fromConfigs
    const pool = HealthyProvidersPool.fromConfigs(
      rpcConfigs,
      createProvider,
      jest.fn() as unknown as Logger
    ) as unknown as ProviderPoolDecorator<InstrumentedEthersProvider>;

    // When
    pool.setProviders("ethereum", providers, providersHealthCheck, 90n);
    const result = pool.getProviders() as unknown as InstrumentedHttpProvider[];

    // Then
    expect(pool).toBeInstanceOf(HealthyProvidersPool);
    expect(JSON.stringify(result[0])).toBe(
      '{"initialDelay":1000,"maxDelay":60000,"retries":0,"timeout":5000,"url":"https://rpc1/v1","chain":"ethereum","health":{"lastRequestDurations":[0.1232]},"logger":{}}'
    );
    expect(JSON.stringify(result[1])).toBe(
      '{"initialDelay":1000,"maxDelay":60000,"retries":0,"timeout":5000,"url":"https://rpc3/v1","chain":"ethereum","health":{"lastRequestDurations":[0.1232]},"logger":{}}'
    );
    expect(JSON.stringify(result[2])).toBe(
      '{"initialDelay":1000,"maxDelay":60000,"retries":0,"timeout":5000,"url":"https://rpc2/v1","chain":"ethereum","health":{"lastRequestDurations":[0.1232]},"logger":{}}'
    );
  });

  it("should set up offline providers", () => {
    // Given
    const rpcConfigs: RpcConfig[] = [
      { url: "https://rpc1/v1", weight: 1 },
      { url: "https://rpc2/v1", weight: 1 },
      { url: "https://rpc3/v1", weight: 1 },
    ];

    const providers = [
      new InstrumentedHttpProvider({ url: "https://rpc1/v1", chain: "ethereum" }),
      new InstrumentedHttpProvider({ url: "https://rpc2/v1", chain: "ethereum" }),
      new InstrumentedHttpProvider({ url: "https://rpc3/v1", chain: "ethereum" }),
    ];

    const providersHealthCheck = [
      {
        isHealthy: true,
        latency: 0.05,
        height: 100n,
        url: "https://rpc1/v1",
      },
      {
        isHealthy: true,
        latency: 0.07,
        height: 100n,
        url: "https://rpc2/v1",
      },
      {
        isHealthy: false,
        latency: undefined,
        height: 100n,
        url: "https://rpc3/v1",
      },
    ];

    // Mock createProvider function
    const createProvider = jest.fn((rpcCfg: RpcConfig) => ({
      getUrl: () => rpcCfg.url,
    })) as unknown as (rpcCfg: RpcConfig) => InstrumentedEthersProvider;

    // Call fromConfigs
    const pool = HealthyProvidersPool.fromConfigs(
      rpcConfigs,
      createProvider,
      jest.fn() as unknown as Logger
    ) as unknown as ProviderPoolDecorator<InstrumentedEthersProvider>;

    // When
    pool.setProviders("ethereum", providers, providersHealthCheck, 90n);
    const result = pool.getProviders() as unknown as InstrumentedHttpProvider[];

    result[2].health.serviceOfflineSince;
    // Then
    expect(pool).toBeInstanceOf(HealthyProvidersPool);
    expect(result[0].health.serviceOfflineSince).toBe(undefined); // Online
    expect(result[1].health.serviceOfflineSince).toBe(undefined); // Online
    expect(result[2].health.serviceOfflineSince).toBeDefined(); // Wed Sep 25 2024 13:04:23 GMT-0300 example
  });
});
