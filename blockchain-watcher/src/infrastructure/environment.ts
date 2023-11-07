import { ChainId, ChainName, Network, toChainName } from "@certusone/wormhole-sdk";
import AbstractWatcher from "./watchers/AbstractWatcher";
import { rootLogger } from "./utils/log";
import winston from "winston";
import EvmWatcher from "./watchers/EvmWatcher";
import AbstractHandler from "./handlers/AbstractHandler";

const MAINNET_RPCS: { [key in ChainName]?: string } = {
  ethereum: process.env.ETH_RPC || "https://rpc.ankr.com/eth",
  bsc: process.env.BSC_RPC || "https://bsc-dataseed2.defibit.io",
  polygon: "https://rpc.ankr.com/polygon",
  avalanche: "https://rpc.ankr.com/avalanche",
  oasis: "https://emerald.oasis.dev",
  algorand: "https://mainnet-api.algonode.cloud",
  fantom: "https://rpc.ankr.com/fantom",
  karura: "https://eth-rpc-karura.aca-api.network",
  acala: "https://eth-rpc-acala.aca-api.network",
  klaytn: "https://klaytn-mainnet-rpc.allthatnode.com:8551",
  celo: "https://forno.celo.org",
  moonbeam: "https://rpc.ankr.com/moonbeam",
  arbitrum: "https://arb1.arbitrum.io/rpc",
  optimism: "https://rpc.ankr.com/optimism",
  aptos: "https://fullnode.mainnet.aptoslabs.com/",
  near: "https://rpc.mainnet.near.org",
  xpla: "https://dimension-lcd.xpla.dev",
  terra2: "https://phoenix-lcd.terra.dev",
  terra: "https://terra-classic-fcd.publicnode.com",
  injective: "https://api.injective.network",
  solana: process.env.SOLANA_RPC ?? "https://api.mainnet-beta.solana.com",
  sui: "https://rpc.mainnet.sui.io",
};

const TESTNET_RPCS: { [key in ChainName]?: string } = {
  bsc: "https://data-seed-prebsc-2-s3.binance.org:8545",
  polygon: "https://matic-mumbai.chainstacklabs.com",
  avalanche: "https://api.avax-test.network/ext/bc/C/rpc",
  celo: "https://alfajores-forno.celo-testnet.org",
  moonbeam: "https://rpc.api.moonbase.moonbeam.network",
};

const DEVNET_RPCS: { [key in ChainName]?: string } = {
  ethereum: "http://localhost:8545",
  bsc: "http://localhost:8546",
};

let hasLoadedDotEnv = false;
function loadDotEnv() {
  if (readEnvironmentVariable("USE_ENV_FILE") === "true" && !hasLoadedDotEnv) {
    //use the dotenv library to load in the .env file
    require("dotenv").config();
    hasLoadedDotEnv = true;
  }
}

const readEnvironmentVariable = (name: string): string | null => {
  const value = process.env[name];
  if (!value) {
    return null;
  }
  return value;
};

export type HandlerConfig = {
  name: string;
  config: any;
};

export type ConfigFile = {
  network: Network;
  supportedChains: ChainId[];
  rpcs: { chain: ChainId; rpc: string }[];
  handlers: HandlerConfig[];
};

export type Environment = {
  network: Network;
  configurationPath: any;
  configuration: ConfigFile;
  supportedChains: ChainId[];
  rpcs: Map<ChainId, string>;
  logger: winston.Logger;
};

let environment: Environment | null = null;

export function getEnvironment(): Environment {
  if (environment) {
    return environment;
  } else {
    throw new Error("Environment not set");
  }
}

export async function initializeEnvironment(configurationPath: string) {
  loadDotEnv();
  const configuration = require(configurationPath);
  const json: ConfigFile = JSON.parse(JSON.stringify(configuration));

  const network = json.network;
  if (network !== "MAINNET" && network !== "TESTNET" && network !== "DEVNET") {
    throw new Error("Invalid network provided in the configuration file");
  }

  const supportedChains = json.supportedChains;
  if (!supportedChains || supportedChains.length === 0) {
    throw new Error("No supported chains provided in the configuration file");
  }

  const configRpcs = json.rpcs;
  const rpcs = new Map<ChainId, string>();
  for (const chain of supportedChains) {
    configRpcs.forEach((item: any) => {
      //double equals for string/int equality
      if (item.chain == chain) {
        if (!item.rpc) {
          throw new Error(`No RPC provided for chain ${chain}`);
        }
        rpcs.set(chain, item.rpc);
      }
    });
  }

  environment = {
    network,
    configurationPath,
    configuration,
    supportedChains,
    rpcs,
    logger: rootLogger,
  };
}

//TODO this
export function createHandlers(env: Environment): AbstractHandler<any>[] {
  const handlerArray: AbstractHandler<any>[] = [];

  for (const handler of env.configuration.handlers) {
    const handlerInstance = new (require(`./handlers/${handler.name}`).default)(
      env,
      handler.config
    );
    handlerArray.push(handlerInstance);
  }

  return handlerArray;
}

//TODO this process probably needs persistence
export function createWatchers(
  env: Environment,
  handlers: AbstractHandler<any>[]
): AbstractWatcher[] {
  const watchers: AbstractWatcher[] = [];
  for (const chain of env.supportedChains) {
    const rpc = env.rpcs.get(chain);
    if (!rpc) {
      throw new Error(`No RPC provided for chain ${chain}`);
    }
    const watcher = new EvmWatcher(
      toChainName(chain) + " Watcher",
      env.network,
      handlers,
      chain,
      rpc,
      env.logger
    );
    watchers.push(watcher);
  }

  return watchers;
}
