import { SnsConfig } from "../../src/infrastructure/repositories";
import { Config, PlatformConfig } from "../../src/infrastructure/config";

export const configMock = (chains: string[] = []): Config => {
  const platformRecord: Record<string, PlatformConfig> = {
    ethereum: {
      name: "ethereum",
      network: "goerli",
      chainId: 2,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    solana: {
      name: "solana",
      network: "devnet",
      chainId: 1,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    karura: {
      name: "karura",
      network: "testnet",
      chainId: 11,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
  };

  const snsConfig: SnsConfig = {
    region: "string",
    topicArn: "string",
    subject: "string",
    groupId: "string",
    credentials: {
      accessKeyId: "string",
      secretAccessKey: "string",
      url: "string",
    },
  };

  const cfg: Config = {
    environment: "testnet",
    port: 999,
    logLevel: "info",
    dryRun: false,
    sns: snsConfig,
    metadata: {
      dir: "./metadata-repo/jobs",
    },
    jobs: {
      dir: "./metadata-repo/jobs",
    },
    platforms: platformRecord,
    supportedChains: chains,
  };

  return cfg;
};
