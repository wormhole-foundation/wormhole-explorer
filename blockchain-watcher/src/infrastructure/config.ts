import config from "config";
import { SnsConfig } from "./repositories/SnsEventRepository";

export type Config = {
  environment: "testnet" | "mainnet";
  port: number;
  logLevel: "debug" | "info" | "warn" | "error";
  dryRun: boolean;
  dbConfig?: DBConfig;
  sns: SnsConfig;
  metadata: {
    use: ("fs" | "postgres")[];
    dir: string;
  };
  jobs: {
    use: ("fs" | "postgres")[];
    dir: string;
  };
  jobExecutions: {
    use: "local" | "postgres";
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

export type DBConfig = {
  connString: string;
  connectionTimeout: number;
  queryTimeout?: number;
  maxPoolSize?: number;
  migrationsDir?: string;
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
  sns: config.get<SnsConfig>("sns"),
  metadata: {
    use: config.get<string[]>("metadata.use") ?? ["fs"],
    dir: config.get<string>("metadata.dir"),
  },
  jobs: {
    use: config.get<string[]>("jobs.use") ?? ["fs"],
    dir: config.get<string>("jobs.dir"),
  },
  jobExecutions: {
    use: config.get<string>("jobExecutions.use") ?? "local",
  },
  chains: config.get<Record<string, ChainRPCConfig>>("chains"),
  enabledPlatforms: config.get<string[]>("enabledPlatforms"),
  dbConfig: config.get<DBConfig>("db"),
} as Config;
