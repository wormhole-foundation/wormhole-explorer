import {
  CHAIN_ID_ACALA,
  CHAIN_ID_ALGORAND,
  CHAIN_ID_APTOS,
  CHAIN_ID_AURORA,
  CHAIN_ID_AVAX,
  CHAIN_ID_BSC,
  CHAIN_ID_CELO,
  CHAIN_ID_ETH,
  CHAIN_ID_FANTOM,
  CHAIN_ID_KARURA,
  CHAIN_ID_KLAYTN,
  CHAIN_ID_MOONBEAM,
  CHAIN_ID_NEAR,
  CHAIN_ID_OASIS,
  CHAIN_ID_POLYGON,
  CHAIN_ID_SOLANA,
  CHAIN_ID_TERRA,
  CHAIN_ID_TERRA2,
  CHAIN_ID_XPLA,
} from "@certusone/wormhole-sdk";

export const WORMHOLE_RPC_HOSTS = [
  "https://wormhole-v2-mainnet-api.certus.one",
  "https://wormhole.inotel.ro",
  "https://wormhole-v2-mainnet-api.mcf.rocks",
  "https://wormhole-v2-mainnet-api.chainlayer.network",
  "https://wormhole-v2-mainnet-api.staking.fund",
  "https://wormhole-v2-mainnet.01node.com",
];

export const CHAIN_ID_MAP = {
  0: undefined,
  1: CHAIN_ID_SOLANA,
  2: CHAIN_ID_ETH,
  3: CHAIN_ID_TERRA,
  4: CHAIN_ID_BSC,
  5: CHAIN_ID_POLYGON,
  6: CHAIN_ID_AVAX,
  7: CHAIN_ID_OASIS,
  8: CHAIN_ID_ALGORAND,
  9: CHAIN_ID_AURORA,
  10: CHAIN_ID_FANTOM,
  11: CHAIN_ID_KARURA,
  12: CHAIN_ID_ACALA,
  13: CHAIN_ID_KLAYTN,
  14: CHAIN_ID_CELO,
  15: CHAIN_ID_NEAR,
  16: CHAIN_ID_MOONBEAM,
  18: CHAIN_ID_TERRA2,
  22: CHAIN_ID_APTOS,
  28: CHAIN_ID_XPLA,
};

import { ethers } from "ethers";
require("dotenv").config();

export const DISALLOWLISTED_ADDRESSES = [
  "0x04132bf45511d03a58afd4f1d36a29d229ccc574",
  "0xa79bd679ce21a2418be9e6f88b2186c9986bbe7d",
  "0x931c3987040c90b6db09981c7c91ba155d3fa31f",
  "0x8fb1a59ca2d57b51e5971a85277efe72c4492983",
  "0xd52d9ba6fcbadb1fe1e3aca52cbb72c4d9bbb4ec",
  "0x1353c55fd2beebd976d7acc4a7083b0618d94689",
  "0xf0fbdb8a402ec0fc626db974b8d019c902deb486",
  "0x1fd4a95f4335cf36cac85730289579c104544328",
  "0x358aa13c52544eccef6b0add0f801012adad5ee3",
  "0xbe32b7acd03bcc62f25ebabd169a35e69ef17601",
  "0x7ffb3d637014488b63fb9858e279385685afc1e2",
  "0x337dc89ebcc33a337307d58a51888af92cfdc81b",
  "0x5Cb89Ac06F34f73B1A6b8000CEb0AfBc97d58B6b",
  "0xd9F0446AedadCf16A12692E02FA26C617FA4D217",
  "0xD7b41531456b636641F7e867eC77120441D1E1E8",
  "0x9f607027b69f6e123bc3bd56a686b735fa75f30a",
  "0x2a35965bbad6fd3964ef815d011c51ab1c546e67",
  "0x053c070f0923a5b770cc59d7bf74ecff991cd0b8",
  "0x3dab0f14ea515d5c842b631bd6df0f7f989c47b3",
  "0x7ee4f716e3c716d61f6158bde3ed5ab03fb6b90c",
  "0x90285e9567be274ae892c88d3ffd76c87d6c7904",
  "0x2d4678e71590c56eb37869832a3642c405e1c252", // fake saitama on poly
  "0x1e49f85f8f5d4ef948ccb953c0172c648b75222f",
  "0x477c7802632f0d38f285a7fd7112a66c11b99db6",
  "0xdaff96cc3d5e2fa982812ec12ce74833deb51327", //fake btc on bsc
  "0xe389ac691bd2b0228daffff548fbce38470373e8", //fake wrapped matic on poly
  "0x7e347498dfef39a88099e3e343140ae17cde260e", //wrapped avax on bsc
  "0x86812b970bbdce75b4590243ba2cbff671d0b754", //fake tether on bsc
  "0x3d8babf3afd0e1bfc843f9638f650fa50ae6c22b", //fake tether on eth
  "0x0749902ae8ed9c6a508271bad18f185dba7185d4", //wrapped eth on polygon
  "0x8e1c62f03b995938233ffa3762bd69f889016b3c", //fake luna2.0 on bsc
].map((x) => x.toLowerCase());

export const COIN_GECKO_EXCEPTIONS = [
  {
    chainId: 2,
    tokenAddress: "0x707f9118e33a9b8998bea41dd0d46f38bb963fc8".toLowerCase(),
    coinGeckoId: "ethereum",
  },
  {
    chainId: 2,
    tokenAddress: "0xdAF566020156297E2837fDfaA6Fbba929A29461E".toLowerCase(),
    coinGeckoId: "safe-coin-2",
  },
  {
    chainId: 2,
    tokenAddress: "0x5ab6A4F46Ce182356B6FA2661Ed8ebcAFce995aD".toLowerCase(),
    coinGeckoId: "sportium",
  },
  {
    chainId: 3,
    tokenAddress: "uluna",
    coinGeckoId: "terra-luna",
  },
  {
    chainId: 3,
    tokenAddress: "ukrw",
    coinGeckoId: "terra-krw",
  },
  {
    chainId: 7,
    tokenAddress: "0x21c718c22d52d0f3a789b752d4c2fd5908a8a733",
    coinGeckoId: "oasis-network", // wrapped-rose does not currently have prices
  },
  {
    chainId: 7,
    tokenAddress: "0x366ef31c8dc715cbeff5fa54ad106dc9c25c6153",
    coinGeckoId: "tether",
  },
  {
    chainId: 1,
    tokenAddress: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    coinGeckoId: "usd-coin",
  },
  {
    chainId: 1,
    tokenAddress: "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
    coinGeckoId: "tether",
  },
  {
    chainId: 11,
    tokenAddress: "0x0000000000000000000500000000000000000007",
    coinGeckoId: "tether",
  },
  {
    chainId: 12,
    tokenAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee2",
    coinGeckoId: "acala",
  },
  {
    chainId: 12,
    tokenAddress: "0x0000000000000000000100000000000000000001",
    coinGeckoId: "acala-dollar",
  },
  {
    chainId: 12,
    tokenAddress: "0x0000000000000000000100000000000000000000",
    coinGeckoId: "acala",
  },
];

export function newProvider(
  url: string,
  batch: boolean = false
): ethers.providers.JsonRpcProvider | ethers.providers.JsonRpcBatchProvider {
  // only support http(s), not ws(s) as the websocket constructor can blow up the entire process
  // it uses a nasty setTimeout(()=>{},0) so we are unable to cleanly catch its errors
  if (url.includes("http")) {
    if (batch) {
      return new ethers.providers.JsonRpcBatchProvider(url);
    }
    return new ethers.providers.JsonRpcProvider(url);
  }
  throw new Error("url does not start with http/https!");
}

export var CHAIN_INFO_MAP = {
  "1": {
    name: "solana",
    evm: false,
    chain_id: CHAIN_ID_SOLANA,
    endpoint_url: process.env.SOLANA_RPC || "https://rpc.ankr.com/solana",
    core_bridge: "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
    token_bridge_address: "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
    custody_address: "GugU1tP7doLeTw9hQP51xRJyS8Da1fWxuiy2rVrnMD2m",
    platform: "solana",
    covalentChain: 1399811149,
  },
  "2": {
    name: "eth",
    evm: true,
    chain_id: CHAIN_ID_ETH,
    endpoint_url: process.env.ETH_RPC || "https://rpc.ankr.com/eth",
    core_bridge: "0x98f3c9e6E3fAce36bAAd05FE09d375Ef1464288B",
    token_bridge_address: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
    custody_address: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
    api_key: process.env.ETHERSCAN_API,
    urlStem: `https://api.etherscan.io`,
    platform: "ethereum",
    covalentChain: 1,
  },
  "3": {
    name: "terra",
    evm: false,
    chain_id: CHAIN_ID_TERRA,
    endpoint_url: "",
    core_bridge: "terra1dq03ugtd40zu9hcgdzrsq6z2z4hwhc9tqk2uy5",
    token_bridge_address: "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
    custody_address: "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
    urlStem: "https://columbus-fcd.terra.dev",
    platform: "terra",
    covalentChain: 3,
  },
  "4": {
    name: "bsc",
    evm: true,
    chain_id: CHAIN_ID_BSC,
    endpoint_url: process.env.BSC_RPC || "https://rpc.ankr.com/bsc	", //moralis_url
    core_bridge: "0x98f3c9e6E3fAce36bAAd05FE09d375Ef1464288B",
    token_bridge_address: "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7",
    custody_address: "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7",
    api_key: process.env.BSCSCAN_API,
    urlStem: `https://api.bscscan.com`,
    platform: "binance-smart-chain",
    covalentChain: 56,
  },
  "5": {
    name: "polygon",
    evm: true,
    chain_id: CHAIN_ID_POLYGON,
    endpoint_url: process.env.POLYGON_RPC || "https://rpc.ankr.com/polygon	",
    core_bridge: "0x7A4B5a56256163F07b2C80A7cA55aBE66c4ec4d7",
    token_bridge_address: "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE",
    custody_address: "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE",
    api_key: process.env.POLYSCAN_API,
    urlStem: `https://api.polygonscan.com`,
    platform: "polygon-pos", //coingecko?,
    covalentChain: 137,
  },
  "6": {
    name: "avalanche",
    evm: true,
    chain_id: CHAIN_ID_AVAX,
    endpoint_url: process.env.AVAX_RPC || "https://rpc.ankr.com/avalanche",
    core_bridge: "0x54a8e5f9c4CbA08F9943965859F6c34eAF03E26c",
    token_bridge_address: "0x0e082F06FF657D94310cB8cE8B0D9a04541d8052",
    custody_address: "0x0e082F06FF657D94310cB8cE8B0D9a04541d8052",
    api_key: process.env.SNOWTRACE_API,
    urlStem: `https://api.snowtrace.io`,
    platform: "avalanche", //coingecko?
    covalentChain: 43114,
  },
  "7": {
    name: "oasis",
    evm: true,
    chain_id: CHAIN_ID_OASIS,
    endpoint_url: "https://emerald.oasis.dev",
    core_bridge: "0xfE8cD454b4A1CA468B57D79c0cc77Ef5B6f64585",
    token_bridge_address: "0x5848C791e09901b40A9Ef749f2a6735b418d7564",
    custody_address: "0x5848C791e09901b40A9Ef749f2a6735b418d7564",
    api_key: "",
    urlStem: `https://explorer.emerald.oasis.dev`,
    platform: "oasis", //coingecko?
    covalentChain: 42262,
  },
  "8": {
    name: "algorand",
    evm: false,
    chain_id: CHAIN_ID_ALGORAND,
    endpoint_url: "https://node.algoexplorerapi.io",
    core_bridge: "842125965",
    token_bridge_address: "842126029",
    custody_address: "842126029",
    api_key: "",
    urlStem: `https://algoexplorer.io`,
    platform: "algorand", //coingecko?
    covalentChain: undefined,
  },
  "9": {
    name: "aurora",
    evm: true,
    chain_id: CHAIN_ID_AURORA,
    endpoint_url: "https://mainnet.aurora.dev",
    core_bridge: "0xa321448d90d4e5b0A732867c18eA198e75CAC48E",
    token_bridge_address: "0x51b5123a7b0F9b2bA265f9c4C8de7D78D52f510F",
    custody_address: "0x51b5123a7b0F9b2bA265f9c4C8de7D78D52f510F",
    api_key: process.env.AURORA_API,
    urlStem: `https://api.aurorascan.dev`, //?module=account&action=txlist&address={addressHash}
    covalentChain: 1313161554,
    platform: "aurora", //coingecko?
  },
  "10": {
    name: "fantom",
    evm: true,
    chain_id: CHAIN_ID_FANTOM,
    endpoint_url: "https://rpc.ftm.tools",
    core_bridge: "0x126783A6Cb203a3E35344528B26ca3a0489a1485",
    token_bridge_address: "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2",
    custody_address: "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2",
    api_key: process.env.FTMSCAN_API,
    urlStem: `https://api.FtmScan.com`,
    platform: "fantom", //coingecko?
    covalentChain: 250,
  },
  "11": {
    name: "karura",
    evm: true,
    chain_id: CHAIN_ID_KARURA,
    endpoint_url: "https://eth-rpc-karura.aca-api.network",
    core_bridge: "0xa321448d90d4e5b0A732867c18eA198e75CAC48E",
    token_bridge_address: "0xae9d7fe007b3327AA64A32824Aaac52C42a6E624",
    custody_address: "0xae9d7fe007b3327AA64A32824Aaac52C42a6E624",
    api_key: "",
    urlStem: `https://blockscout.karura.network`,
    platform: "karura", //coingecko?
    covalentChain: "",
  },
  "12": {
    name: "acala",
    evm: true,
    chain_id: CHAIN_ID_ACALA,
    endpoint_url: "https://eth-rpc-acala.aca-api.network",
    core_bridge: "0xa321448d90d4e5b0A732867c18eA198e75CAC48E",
    token_bridge_address: "0xae9d7fe007b3327AA64A32824Aaac52C42a6E624",
    custody_address: "0xae9d7fe007b3327AA64A32824Aaac52C42a6E624",
    api_key: "",
    urlStem: `https://blockscout.acala.network`,
    platform: "acala", //coingecko?
    covalentChain: "",
  },
  "13": {
    name: "klaytn",
    evm: true,
    chain_id: CHAIN_ID_KLAYTN,
    endpoint_url: "https://klaytn-mainnet-rpc.allthatnode.com:8551",
    core_bridge: "0x0C21603c4f3a6387e241c0091A7EA39E43E90bb7",
    token_bridge_address: "0x5b08ac39EAED75c0439FC750d9FE7E1F9dD0193F",
    custody_address: "0x5b08ac39EAED75c0439FC750d9FE7E1F9dD0193F",
    api_key: process.env.KLAYTN_API,
    urlStem: `https://scope.klaytn.com`,
    platform: "klay-token", //coingecko?
    covalentChain: "8217",
  },
  "14": {
    name: "celo",
    evm: true,
    chain_id: CHAIN_ID_CELO,
    endpoint_url: "https://forno.celo.org",
    core_bridge: "0xa321448d90d4e5b0A732867c18eA198e75CAC48E",
    token_bridge_address: "0x796Dff6D74F3E27060B71255Fe517BFb23C93eed",
    custody_address: "0x796Dff6D74F3E27060B71255Fe517BFb23C93eed",
    api_key: "",
    urlStem: `https://explorer.celo.org`,
    platform: "celo", //coingecko?
    covalentChain: "42220",
  },
  "15": {
    name: "near",
    evm: false,
    chain_id: CHAIN_ID_NEAR,
    endpoint_url: "https://rpc.ankr.com/near",
    core_bridge: "contract.wormhole_crypto.near",
    token_bridge_address: "contract.portalbridge.near",
    custody_address: "contract.portalbridge.near",
    urlStem: "",
    platform: "near",
    covalentChain: undefined,
  },
  "16": {
    name: "moonbeam",
    evm: true,
    chain_id: CHAIN_ID_MOONBEAM,
    endpoint_url: "https://rpc.api.moonbeam.network",
    core_bridge: "0xC8e2b0cD52Cf01b0Ce87d389Daa3d414d4cE29f3",
    token_bridge_address: "0xB1731c586ca89a23809861c6103F0b96B3F57D92",
    custody_address: "0xB1731c586ca89a23809861c6103F0b96B3F57D92",
    api_key: process.env.MOONBEAM_API,
    urlStem: "https://api-moonbeam.moonscan.io",
    platform: "moonbeam",
    covalentChain: 1284,
  },
  "18": {
    name: "terra2",
    evm: false,
    chain_id: CHAIN_ID_TERRA2,
    endpoint_url: "",
    core_bridge:
      "terra12mrnzvhx3rpej6843uge2yyfppfyd3u9c3uq223q8sl48huz9juqffcnhp",
    token_bridge_address:
      "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
    custody_address:
      "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
    urlStem: "https://phoenix-fcd.terra.dev",
    platform: "terra",
    covalentChain: 3,
  },
  "22": {
    name: "aptos",
    evm: false,
    chain_id: CHAIN_ID_APTOS,
    endpoint_url: "https://fullnode.mainnet.aptoslabs.com/v1",
    core_bridge:
      "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
    token_bridge_address:
      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
    custody_address:
      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
    api_key: "",
    urlStem: "",
    platform: "aptos",
    covalentChain: 0,
  },
};

export async function sleepFor(timeInMs: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, timeInMs);
  });
}
