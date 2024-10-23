import { SnsConfig, SqsConfig } from "./repositories/target/SnsEventRepository";
import { InfluxConfig } from "./repositories/target/InfluxEventRepository";
import config from "config";

export type Environment = "testnet" | "mainnet";

export type LogLevel = "debug" | "info" | "warn" | "error";

export type Config = {
  environment: Environment;
  port: number;
  logLevel: LogLevel;
  dryRun: boolean;
  rpcHealthcheckInterval: number;
  sns: SnsConfig;
  sqs: SqsConfig;
  influx?: InfluxConfig;
  metadata?: {
    dir: string;
  };
  jobs: {
    dir: string;
  };
  chains: Record<string, ChainRPCConfig>;
  enabledPlatforms: string[];
};

export type ChainRPCConfig = {
  name: string;
  network: string;
  chainId: number;
  rpcs: string[];
  timeout?: number;
  retries?: number;
  rateLimit?: {
    period: number;
    limit: number;
  };
};

/*
  By setting NODE_CONFIG_ENV we can point to a different config directory.
  Default settings can be customized by definining NODE_ENV=staging|production.
  Some options may be overridable by env variables, see: config/custom-environment-variables.json

  For array values, you should use something like this:
  ETHEREUM_RPCS='["http://1.com","http://2.com"]'
*/
export const configuration = {
  environment: config.get<string>("environment"),
  port: config.get<number>("port") ?? 9090,
  logLevel: config.get<string>("logLevel")?.toLowerCase() ?? "info",
  dryRun: config.get<string>("dryRun") === "true" ? true : false,
  rpcHealthcheckInterval: config.get<number>("rpcHealthcheckInterval") ?? 600000,
  sns: config.get<SnsConfig>("sns"),
  sqs: config.get<SqsConfig>("sqs"),
  influx: config.get<InfluxConfig>("influx"),
  metadata: {
    dir: config.get<string>("metadata.dir"),
  },
  jobs: {
    dir: config.get<string>("jobs.dir"),
  },
  chains: config.get<Record<string, ChainRPCConfig>>("chains"),
  enabledPlatforms: config.get<string[]>("enabledPlatforms"),
} as Config;
