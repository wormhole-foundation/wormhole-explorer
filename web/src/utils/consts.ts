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
