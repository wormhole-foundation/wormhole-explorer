import { jest } from "@jest/globals";
import axios from "axios";

export class ProviderHealthInstrumentationMock {
  fetch = async (input: string | URL | Request, init?: RequestInit) => {
    const body = typeof init?.body === 'string' ? JSON.parse(init.body) : init?.body;
    const res = await axios.request({
      url: input.toString(),
      method: 'POST',
      data: body
    });
    return {
      status: 200,
      json: () => res.data
    };
  }
}

type RpcConfig = {url:string};
type PoolSupplier = <T>(cfg: RpcConfig, createProvider: (cfg: RpcConfig) => T, type?: string) => { get: () => T };
const providerPoolSupplier: PoolSupplier = <T>(cfg: RpcConfig, createProvider: (cfg: RpcConfig) => T, type?: string) => {
  return {
    get: () => createProvider(cfg)
  };
}

export function mockRpcPool() {
  jest.mock('@xlabs/rpc-pool', () => {
    return {
      ProviderHealthInstrumentation: ProviderHealthInstrumentationMock,
      providerPoolSupplier
    };
  });
}