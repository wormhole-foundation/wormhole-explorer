import {
  CHAINS,
  CHAIN_ID_TO_NAME,
  ChainId,
  ChainName,
  Network,
} from "@certusone/wormhole-sdk";
import { getWormholeRelayerAddress } from "@certusone/wormhole-sdk/lib/cjs/relayer";

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

let Environment: Network | undefined;

export async function getEnvironment(): Promise<Network> {
  if (Environment) {
    return Environment;
  }
  loadDotEnv();
  const environment = readEnvironmentVariable("ENVIRONMENT");
  console.log(`Environment: ${environment}`);
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
      //get wormhole relayer address throws an error if the address isn't found
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
  const network = await getEnvironment();
  const SUPPORTED_CHAINS = await getSupportedChains();
  const rpcs = new Map<ChainId, string>();
  const rpcsString = readEnvironmentVariable("RPCS");
  const rpcObject = rpcsString ? JSON.parse(rpcsString) : {};
  const useDefaultRpcs = readEnvironmentVariable("USE_DEFAULT_RPCS");
  for (const chainId of SUPPORTED_CHAINS) {
    if (rpcObject[chainId]) {
      rpcs.set(chainId, rpcObject[chainId]);
    } else {
      const defaultRpc = getDefaultRpc(network, chainId);
      if (defaultRpc && useDefaultRpcs === "true") {
        rpcs.set(chainId, defaultRpc);
      } else {
        throw new Error(`RPC not found for chain ${chainId}`);
      }
    }
  }
  return rpcs;
}

function getDefaultRpc(network: Network, chainId: ChainId): string | null {
  const chainName = CHAIN_ID_TO_NAME[chainId];
  if (network === "MAINNET") {
    return MAINNET_RPCS[chainName] ?? null;
  } else if (network === "TESTNET") {
    return TESTNET_RPCS[chainName] ?? null;
  } else if (network === "DEVNET") {
    return DEVNET_RPCS[chainName] ?? null;
  }
  return null;
}
