import {
  CHAINS,
  CHAIN_ID_TO_NAME,
  ChainId,
  ChainName,
  Network,
} from "@certusone/wormhole-sdk";
import { getWormholeRelayerAddress } from "@certusone/wormhole-sdk/lib/cjs/relayer";

let hasLoadedDotEnv = false;
function loadDotEnv() {
  if (readEnvironmentVariable("USE_ENV_FILE") === "true" && !hasLoadedDotEnv) {
    //use the dotenv library to load in the .env file
    require("dotenv").config();
    hasLoadedDotEnv = true;
  }
}

const readEnvironmentVariable = (name: string): string => {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Environment variable ${name} is not set`);
  }
  return value;
};

let Environment: Network | undefined;

export async function getEnvironment(): Promise<Network> {
  if (Environment) {
    return Environment;
  }
  loadDotEnv();
  const environment = readEnvironmentVariable("ENVIRONMENT");
  switch (environment) {
    case "MAINNET":
      Environment = "MAINNET";
      return "MAINNET";
    case "TESTNET":
      Environment = "TESTNET";
      return "TESTNET";
    case "DEVNET":
      Environment = "DEVNET";
      return "DEVNET";
    default:
      throw new Error(`Unknown environment ${environment}`);
  }
}

export function getWormholeRelayerAddressWrapped(
  chain: ChainName,
  network: Network
): string {
  loadDotEnv();
  const address = readEnvironmentVariable(
    `${chain.toUpperCase()}_${network}_WORMHOLE_RELAYER_ADDRESS`
  );
  if (address) {
    return address;
  } else {
    return getWormholeRelayerAddress(chain, network);
  }
  return address;
}

export async function getSupportedChains(): Promise<ChainId[]> {
  const ENVIRONMENT = await getEnvironment();
  const filteredChains = Object.values(CHAINS).filter((chain) => {
    let address: string | undefined;
    try {
      address = getWormholeRelayerAddressWrapped(
        CHAIN_ID_TO_NAME[chain],
        ENVIRONMENT
      );
    } catch (e) {
      address = undefined;
    }

    if (address) {
      return true;
    } else {
      return false;
    }
  });
  return filteredChains.map((chain) => {
    return chain as ChainId;
  });
}

export async function getRpcs(): Promise<Map<ChainId, string>> {
  loadDotEnv();
  const SUPPORTED_CHAINS = await getSupportedChains();
  const rpcs = new Map<ChainId, string>();
  const rpcsString = readEnvironmentVariable("RPCS");
  const rpcObject = JSON.parse(rpcsString);
  for (const chainId of SUPPORTED_CHAINS) {
    if (rpcObject[chainId]) {
      rpcs.set(chainId, rpcObject[chainId]);
    } else {
      throw new Error(`RPC not found for chain ${chainId}`);
    }
  }
  return rpcs;
}
