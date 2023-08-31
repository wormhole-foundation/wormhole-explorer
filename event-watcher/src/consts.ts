import { ChainName, CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { AxiosRequestConfig } from 'axios';
import { env } from './config';

export const TIMEOUT = 0.5 * 1000;

// Notes about RPCs
// Ethereum
//   ethereum: "https://rpc.ankr.com/eth", // "finalized" does not work on Ankr as of 2022-12-16
// BSC
//   https://docs.bscscan.com/misc-tools-and-utilities/public-rpc-nodes
//   bsc: "https://bsc-dataseed1.binance.org", // Cannot read properties of undefined (reading 'error')
//   'https://rpc.ankr.com/bsc' has been very slow, trying a diff rpc
// Avalanche
//   https://docs.avax.network/apis/avalanchego/public-api-server
//   avalanche: "https://api.avax.network/ext/bc/C/rpc", // 500 error on batch request
// Fantom
//   fantom: "https://rpc.ftm.tools", // Cannot read properties of null (reading 'timestamp')"
// Klaytn
// this one immediately 429s
// klaytn: 'https://public-node-api.klaytnapi.com/v1/cypress',
// Near
//   archive node
//   https://archival-rpc.mainnet.near.org
// Arbitrum
//  This node didn't work:  'https://arb1.arbitrum.io/rpc',

export const RPCS_BY_CHAIN: { [key in ChainName]?: string } = {
  acala: 'https://eth-rpc-acala.aca-api.network',
  algorand: 'https://mainnet-api.algonode.cloud',
  aptos: 'https://fullnode.mainnet.aptoslabs.com/',
  arbitrum: 'https://arb1.arbitrum.io/rpc',
  avalanche: 'https://rpc.ankr.com/avalanche',
  base: 'https://developer-access-mainnet.base.org',
  bsc: 'https://bsc-dataseed2.defibit.io',
  celo: 'https://forno.celo.org',
  ethereum: env.ETH_RPC || 'https://svc.blockdaemon.com/ethereum/mainnet/native',
  fantom: 'https://rpc.ankr.com/fantom',
  injective: 'https://api.injective.network',
  karura: 'https://eth-rpc-karura.aca-api.network',
  klaytn: "https://public-en-cypress.klaytn.net", // 'https://klaytn-mainnet-rpc.allthatnode.com:8551',
  moonbeam: 'https://rpc.ankr.com/moonbeam',
  near: 'https://rpc.mainnet.near.org',
  oasis: 'https://emerald.oasis.dev',
  optimism: 'https://rpc.ankr.com/optimism',
  polygon: 'https://rpc.ankr.com/polygon',
  solana: 'https://api.mainnet-beta.solana.com',
  sui: 'https://rpc.mainnet.sui.io',
  terra: 'https://terra-classic-fcd.publicnode.com', // 'https://columbus-fcd.terra.dev',
  terra2: 'https://phoenix-lcd.terra.dev',
  xpla: 'https://dimension-lcd.xpla.dev',
  sei: 'https://sei-rest.brocha.in', // https://docs.sei.io/develop/resources
};

// Separating for now so if we max out infura we can keep Polygon going
export const POLYGON_ROOT_CHAIN_RPC = 'https://rpc.ankr.com/eth';
export const POLYGON_ROOT_CHAIN_ADDRESS = '0x86E4Dc95c7FBdBf52e33D563BbDB00823894C287';
// Optimism watcher relies on finalized calls which don't work right on Ankr
export const OPTIMISM_CTC_CHAIN_RPC = env.ETH_RPC;
export const OPTIMISM_CTC_CHAIN_ADDRESS = '0x5E4e65926BA27467555EB562121fac00D24E9dD2';

export const ALGORAND_INFO = {
  appid: Number(CONTRACTS.MAINNET.algorand.core),
  algodToken: '',
  algodServer: RPCS_BY_CHAIN.algorand,
  algodPort: 443,
  server: 'https://mainnet-idx.algonode.cloud',
  port: 443,
  token: '',
};

export const SEI_EXPLORER_GRAPHQL = 'https://pacific-1-graphql.alleslabs.dev/v1/graphql';
export const SEI_EXPLORER_TXS = 'https://celatone-api.alleslabs.dev/txs/sei/pacific-1/';

// without this, axios request will error `Z_BUF_ERROR`: https://github.com/axios/axios/issues/5346
export const AXIOS_CONFIG_JSON: AxiosRequestConfig = {
  headers: {
    'Accept-Encoding': 'application/json',
    Authorization: 'Bearer zpka_213d294a9a5a44619cd6a02e55a20417_5f43e4d0',
  },
};

export const GUARDIAN_RPC_HOSTS = [
  'https://wormhole-v2-mainnet-api.certus.one',
  'https://wormhole.inotel.ro',
  'https://wormhole-v2-mainnet-api.mcf.rocks',
  'https://wormhole-v2-mainnet-api.chainlayer.network',
  'https://wormhole-v2-mainnet-api.staking.fund',
];
