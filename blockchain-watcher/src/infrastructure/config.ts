import config from "config";
import { SnsConfig } from "./repositories/SnsEventRepository";

export type Config = {
  environment: "testnet" | "mainnet";
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

// By setting NODE_CONFIG_ENV we can point to a different config directory.
// Default settings can be customized by definining NODE_ENV=staging|production.
export const configuration = {
  environment: config.get<string>("environment"),
  dryRun: config.get<string>("dryRun") === "true" ? true : false,
  sns: config.get<SnsConfig>("sns"),
  metadata: {
    dir: config.get<string>("metadata.dir"),
  },
  platforms: config.get<Record<string, PlatformConfig>>("platforms"),
  supportedChains: config.get<string[]>("supportedChains"),
} as Config;
