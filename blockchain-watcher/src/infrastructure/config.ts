import config from "config";
import { SnsConfig } from "./repositories/SnsEventRepository";

export type Config = {
  environment: "testnet" | "mainnet";
  port: number;
  logLevel: "debug" | "info" | "warn" | "error";
  dryRun: boolean;
  sns: SnsConfig;
  metadata?: {
    dir: string;
  };
  platforms: Record<string, PlatformConfig>;
  supportedChains: string[];
};

export type PlatformConfig = {
  name: string;
  network: string;
  chainId: number;
  rpcs: string[];
  timeout?: number;
};

/*
  By setting NODE_CONFIG_ENV we can point to a different config directory.
  Default settings can be customized by definining NODE_ENV=staging|production.
  Some options may be overridable by env variables, see: config/custom-environment-variables.json
*/
export const configuration = {
  environment: config.get<string>("environment"),
  port: config.get<number>("port") ?? 9090,
  logLevel: config.get<string>("logLevel")?.toLowerCase() ?? "info",
  dryRun: config.get<string>("dryRun") === "true" ? true : false,
  sns: config.get<SnsConfig>("sns"),
  metadata: {
    dir: config.get<string>("metadata.dir"),
  },
  platforms: config.get<Record<string, PlatformConfig>>("platforms"),
  supportedChains: config.get<string[]>("supportedChains"),
} as Config;
