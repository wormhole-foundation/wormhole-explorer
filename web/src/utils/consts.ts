import {
  CHAIN_ID_SOLANA,
  CHAIN_ID_ETH,
  CHAIN_ID_TERRA,
  CHAIN_ID_BSC,
  CHAIN_ID_POLYGON,
  CHAIN_ID_AVAX,
  CHAIN_ID_OASIS,
  CHAIN_ID_ALGORAND,
  CHAIN_ID_AURORA,
  CHAIN_ID_FANTOM,
  CHAIN_ID_KARURA,
  CHAIN_ID_ACALA,
  CHAIN_ID_KLAYTN,
  CHAIN_ID_CELO,
  CHAIN_ID_TERRA2,
  ChainId,
  CHAIN_ID_NEAR,
  CHAIN_ID_MOONBEAM,
  CHAIN_ID_UNSET,
} from "@certusone/wormhole-sdk";

require("dotenv").config();

export const EXPECTED_GUARDIAN_COUNT = 19;
export const POLL_TIME = 60 * 1000;

export type CHAIN_INFO = {
  name: string;
  evm: boolean;
  chainId: ChainId;
  endpointUrl: any;
  platform: string;
  covalentChain: number;
  explorerStem: string;
  apiKey: string;
  urlStem: string;
};

export const CHAIN_INFO_MAP: { [key: string]: CHAIN_INFO } = {
  1: {
    name: "solana",
    evm: false,
    chainId: CHAIN_ID_SOLANA,
    urlStem: `https://public-api.solscan.io`,
    endpointUrl:
      process.env.REACT_APP_SOLANA_RPC || "https://api.mainnet-beta.solana.com",
    apiKey: "",
    platform: "solana",
    covalentChain: 1399811149,
    explorerStem: `https://solscan.io`,
  },
  2: {
    name: "eth",
    evm: true,
    chainId: CHAIN_ID_ETH,
    endpointUrl: process.env.REACT_APP_ETH_RPC || "https://rpc.ankr.com/eth",
    apiKey: "",
    urlStem: `https://api.etherscan.io`,
    platform: "ethereum",
    covalentChain: 1,
    explorerStem: `https://etherscan.io`,
  },
  3: {
    name: "terra",
    evm: false,
    chainId: CHAIN_ID_TERRA,
    endpointUrl: "",
    apiKey: "",
    urlStem: "https://columbus-fcd.terra.dev",
    platform: "terra",
    covalentChain: 3,
    explorerStem: `https://finder.terra.money/classic`,
  },
  4: {
    name: "bsc",
    evm: true,
    chainId: CHAIN_ID_BSC,
    endpointUrl:
      process.env.REACT_APP_BSC_RPC || "https://bsc-dataseed2.defibit.io",
    apiKey: "",
    urlStem: `https://api.bscscan.com`,
    platform: "binance-smart-chain",
    covalentChain: 56,
    explorerStem: `https://bscscan.com`,
  },
  5: {
    name: "polygon",
    evm: true,
    chainId: CHAIN_ID_POLYGON,
    endpointUrl: process.env.REACT_APP_POLYGON_RPC || "https://polygon-rpc.com",
    apiKey: "",
    urlStem: `https://api.polygonscan.com`,
    platform: "polygon-pos", //coingecko?,
    covalentChain: 137,
    explorerStem: `https://polygonscan.com`,
  },
  6: {
    name: "avalanche",
    evm: true,
    chainId: CHAIN_ID_AVAX,
    endpointUrl:
      process.env.REACT_APP_AVAX_RPC || "https://api.avax.network/ext/bc/C/rpc",
    apiKey: "",
    urlStem: `https://api.snowtrace.io`,
    platform: "avalanche", //coingecko?
    covalentChain: 43114,
    explorerStem: `https://snowtrace.io`,
  },
  7: {
    name: "oasis",
    evm: true,
    chainId: CHAIN_ID_OASIS,
    endpointUrl: "https://emerald.oasis.dev",
    apiKey: "",
    urlStem: `https://explorer.emerald.oasis.dev`,
    platform: "oasis", //coingecko?
    covalentChain: 0,
    explorerStem: `https://explorer.emerald.oasis.dev`,
  },
  8: {
    name: "algorand",
    evm: false,
    chainId: CHAIN_ID_ALGORAND,
    endpointUrl: "https://node.algoexplorerapi.io",
    apiKey: "",
    urlStem: `https://algoexplorer.io`,
    platform: "algorand", //coingecko?
    covalentChain: 0,
    explorerStem: `https://algoexplorer.io`,
  },
  9: {
    name: "aurora",
    evm: true,
    chainId: CHAIN_ID_AURORA,
    endpointUrl: "https://mainnet.aurora.dev",
    apiKey: "",
    urlStem: `https://api.aurorascan.dev`, //?module=account&action=txlist&address={addressHash}
    covalentChain: 1313161554,
    platform: "aurora", //coingecko?
    explorerStem: `https://aurorascan.dev`,
  },
  10: {
    name: "fantom",
    evm: true,
    chainId: CHAIN_ID_FANTOM,
    endpointUrl: "https://rpc.ftm.tools",
    apiKey: "",
    urlStem: `https://api.FtmScan.com`,
    platform: "fantom", //coingecko?
    covalentChain: 250,
    explorerStem: `https://ftmscan.com`,
  },
  11: {
    name: "karura",
    evm: true,
    chainId: CHAIN_ID_KARURA,
    endpointUrl: "https://eth-rpc-karura.aca-api.network",
    apiKey: "",
    urlStem: `https://blockscout.karura.network`,
    platform: "karura", //coingecko?
    covalentChain: 0,
    explorerStem: `https://blockscout.karura.network`,
  },
  12: {
    name: "acala",
    evm: true,
    chainId: CHAIN_ID_ACALA,
    endpointUrl: "https://eth-rpc-acala.aca-api.network",
    apiKey: "",
    urlStem: `https://blockscout.acala.network`,
    platform: "acala", //coingecko?
    covalentChain: 0,
    explorerStem: `https://blockscout.acala.network`,
  },
  13: {
    name: "klaytn",
    evm: true,
    chainId: CHAIN_ID_KLAYTN,
    endpointUrl: "https://klaytn-mainnet-rpc.allthatnode.com:8551",
    apiKey: "",
    urlStem: "https://api-cypress-v2.scope.klaytn.com/v2" || "",
    platform: "klay-token", //coingecko?
    covalentChain: 8217,
    explorerStem: `https://scope.klaytn.com`,
  },
  14: {
    name: "celo",
    evm: true,
    chainId: CHAIN_ID_CELO,
    endpointUrl: "https://forno.celo.org",
    apiKey: "",
    urlStem: `https://explorer.celo.org`,
    platform: "celo", //coingecko?
    covalentChain: 0,
    explorerStem: `https://explorer.celo.org`,
  },
  15: {
    name: "near",
    evm: false,
    chainId: CHAIN_ID_NEAR,
    endpointUrl: "",
    apiKey: "",
    urlStem: `https://explorer.near.org`,
    platform: "near", //coingecko?
    covalentChain: 0,
    explorerStem: `https://explorer.near.org`,
  },
  16: {
    name: "moonbeam",
    evm: true,
    chainId: CHAIN_ID_MOONBEAM,
    endpointUrl: "https://rpc.ankr.com/moonbeam",
    apiKey: "",
    urlStem: `https://api-moonbeam.moonscan.io`,
    platform: "moonbeam", //coingecko?
    covalentChain: 0,
    explorerStem: `https://moonscan.io/`,
  },
  18: {
    name: "terra2",
    evm: false,
    chainId: CHAIN_ID_TERRA2,
    endpointUrl: "",
    apiKey: "",
    urlStem: "https://phoenix-fcd.terra.dev",
    platform: "terra",
    covalentChain: 3,
    explorerStem: `https://finder.terra.money/mainnet`,
  },
};

export const WORMHOLE_RPC_HOSTS = [
  "https://wormhole-v2-mainnet-api.certus.one",
  "https://wormhole.inotel.ro",
  "https://wormhole-v2-mainnet-api.mcf.rocks",
  "https://wormhole-v2-mainnet-api.chainlayer.network",
  "https://wormhole-v2-mainnet-api.staking.fund",
  "https://wormhole-v2-mainnet.01node.com",
];

export const CHAIN_ID_MAP = {
  "1": CHAIN_ID_SOLANA,
  "2": CHAIN_ID_ETH,
  "3": CHAIN_ID_TERRA,
  "4": CHAIN_ID_BSC,
  "5": CHAIN_ID_POLYGON,
  "6": CHAIN_ID_AVAX,
  "7": CHAIN_ID_OASIS,
  "8": CHAIN_ID_ALGORAND,
  "9": CHAIN_ID_AURORA,
  "10": CHAIN_ID_FANTOM,
  "11": CHAIN_ID_KARURA,
  "12": CHAIN_ID_ACALA,
  "13": CHAIN_ID_KLAYTN,
  "14": CHAIN_ID_CELO,
  "15": CHAIN_ID_NEAR,
  "16": CHAIN_ID_MOONBEAM,
  "18": CHAIN_ID_TERRA2,
};

export function findChainId(chain: string) {
  if (chain === "1") {
    return CHAIN_ID_SOLANA;
  } else if (chain === "2") {
    return CHAIN_ID_ETH;
  } else if (chain === "3") {
    return CHAIN_ID_TERRA;
  } else if (chain === "4") {
    return CHAIN_ID_BSC;
  } else if (chain === "5") {
    return CHAIN_ID_POLYGON;
  } else if (chain === "6") {
    return CHAIN_ID_AVAX;
  } else if (chain === "7") {
    return CHAIN_ID_OASIS;
  } else if (chain === "8") {
    return CHAIN_ID_ALGORAND;
  } else if (chain === "9") {
    return CHAIN_ID_AURORA;
  } else if (chain === "10") {
    return CHAIN_ID_FANTOM;
  } else if (chain === "11") {
    return CHAIN_ID_KARURA;
  } else if (chain === "12") {
    return CHAIN_ID_ACALA;
  } else if (chain === "13") {
    return CHAIN_ID_KLAYTN;
  } else if (chain === "14") {
    return CHAIN_ID_CELO;
  } else if (chain === "15") {
    return CHAIN_ID_NEAR;
  } else if (chain === "16") {
    return CHAIN_ID_MOONBEAM;
  } else if (chain === "18") {
    return CHAIN_ID_TERRA2;
  } else {
    return CHAIN_ID_UNSET;
  }
}
