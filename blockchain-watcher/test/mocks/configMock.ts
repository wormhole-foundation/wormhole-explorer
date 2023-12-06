import { SnsConfig } from "../../src/infrastructure/repositories";
import { Config, ChainRPCConfig } from "../../src/infrastructure/config";

export const configMock = (chains: string[] = []): Config => {
  const chainsRecord: Record<string, ChainRPCConfig> = {
    solana: {
      name: "solana",
      network: "devnet",
      chainId: 1,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    ethereum: {
      name: "ethereum",
      network: "goerli",
      chainId: 2,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    bsc: {
      name: "bsc",
      network: "BNB Smart Chain testnet",
      chainId: 4,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    avalanche: {
      name: "avalanche",
      network: "testnet",
      chainId: 6,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    oasis: {
      name: "oasis",
      network: "emerald",
      chainId: 7,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    fantom: {
      name: "fantom",
      network: "testnet",
      chainId: 10,
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
    acala: {
      name: "acala",
      network: "testnet",
      chainId: 12,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    klaytn: {
      name: "klaytn",
      network: "baobab",
      chainId: 13,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    celo: {
      name: "celo",
      network: "alfajores",
      chainId: 14,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    optimism: {
      name: "optimism",
      network: "goerli",
      chainId: 12,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
    base: {
      name: "base",
      network: "goerli",
      chainId: 12,
      rpcs: ["http://localhost"],
      timeout: 10000,
    },
  };

  const snsConfig: SnsConfig = {
    region: "us-east",
    topicArn: "123333223232s",
    subject: "",
    groupId: "1",
    credentials: {
      accessKeyId: "212312312323",
      secretAccessKey: "244122wdsd",
      url: "",
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
    chains: chainsRecord,
    supportedChains: chains,
    enabledChains: ["solana", "evm"],
  };

  return cfg;
};
