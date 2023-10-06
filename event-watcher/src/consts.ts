import { ChainName, CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { AxiosRequestConfig } from 'axios';
import { env } from './config';

export const NETWORK: 'mainnet' | 'testnet' = env.P2P_NETWORK === 'testnet' ? 'testnet' : 'mainnet';
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

export const RPCS_BY_CHAIN_MAINNET: { [key in ChainName]?: string } = {
  acala: env.ACALA_RPC || 'https://eth-rpc-acala.aca-api.network',
  algorand: env.ALGORAND_RPC || 'https://mainnet-api.algonode.cloud',
  aptos: env.APTOS_RPC || 'https://fullnode.mainnet.aptoslabs.com',
  arbitrum: env.ARBITRUM_RPC || 'https://arb1.arbitrum.io/rpc',
  avalanche: env.AVALANCHE_RPC || 'https://rpc.ankr.com/avalanche',
  base: env.BASE_RPC || 'https://rpc.ankr.com/base',
  bsc: env.BSC_RPC || 'https://rpc.ankr.com/bsc_testnet_chapel',
  celo: env.CELO_RPC || 'https://forno.celo.org',
  ethereum: env.ETHEREUM_RPC || 'https://eth.llamarpc.com', // 'https://svc.blockdaemon.com/ethereum/mainnet/native',
  fantom: env.FANTOM_RPC || 'https://rpc.ankr.com/fantom',
  injective: env.INJECTIVE_RPC || 'https://api.injective.network',
  karura: env.KARURA_RPC || 'https://eth-rpc-karura.aca-api.network',
  klaytn: env.KLAYTN_RPC || 'https://public-en-cypress.klaytn.net', // 'https://klaytn-mainnet-rpc.allthatnode.com:8551',
  moonbeam: env.MOONBEAM_RPC || 'https://rpc.ankr.com/moonbeam',
  near: env.NEAR_RPC || 'https://rpc.mainnet.near.org', // 'https://archival-rpc.mainnet.near.org' (older than 5 epochs or ~2.5 days) -> 'https://rpc.mainnet.near.org' [https://stackoverflow.com/questions/66839103/unable-to-get-near-protocol-transaction-status-via-rpc/67199078#67199078]
  oasis: env.OASIS_RPC || 'https://emerald.oasis.dev',
  optimism: env.OPTIMISM_RPC || 'https://rpc.ankr.com/optimism',
  polygon: env.POLYGON_RPC || 'https://rpc.ankr.com/polygon',
  sei: env.SEI_RPC || 'https://sei-rest.brocha.in', // https://docs.sei.io/develop/resources
  solana: env.SOLANA_RPC || 'https://api.mainnet-beta.solana.com',
  sui: env.SUI_RPC || 'https://rpc.mainnet.sui.io',
  terra: env.TERRA_RPC || 'https://terra-classic-fcd.publicnode.com', // 'https://columbus-fcd.terra.dev',
  terra2: env.TERRA2_RPC || 'https://phoenix-lcd.terra.dev',
  wormchain: env.WORMCHAIN_RPC || 'https://wormchain-rpc.quickapi.com',
  xpla: env.XPLA_RPC || 'https://dimension-lcd.xpla.dev',
} as const;

export const RPCS_BY_CHAIN_TESTNET: { [key in ChainName]?: string } = {
  acala: env.ACALA_RPC || 'https://eth-rpc-acala-testnet.aca-staging.network',
  algorand: env.ALGORAND_RPC || 'https://testnet-api.algonode.cloud',
  aptos: env.APTOS_RPC || 'https://aptos-testnet-rpc.allthatnode.com/v1',
  arbitrum: env.ARBITRUM_RPC || 'https://goerli-rollup.arbitrum.io/rpc',
  avalanche: env.AVALANCHE_RPC || 'https://api.avax-test.network/ext/bc/C/rpc',
  base: env.BASE_RPC || 'https://goerli.base.org',
  bsc: env.BSC_RPC || 'https://rpc.ankr.com/bsc_testnet_chapel',
  celo: env.CELO_RPC || 'https://alfajores-forno.celo-testnet.org',
  ethereum: env.ETHEREUM_RPC || 'https://rpc.ankr.com/eth_goerli',
  fantom: env.FANTOM_RPC || 'https://rpc.ankr.com/fantom_testnet',
  injective: env.INJECTIVE_RPC || 'https://testnet.sentry.tm.injective.network:443',
  karura: env.KARURA_RPC || 'https://eth-rpc-karura-testnet.aca-staging.network',
  klaytn: env.KLAYTN_RPC || 'https://api.baobab.klaytn.net:8651',
  moonbeam: env.MOONBEAM_RPC || 'https://rpc.api.moonbase.moonbeam.network',
  near: env.NEAR_RPC || 'https://testnet.aurora.dev',
  oasis: env.OASIS_RPC || 'https://testnet.emerald.oasis.dev',
  optimism: env.OPTIMISM_RPC || 'https://optimism-goerli.publicnode.com',
  polygon: env.POLYGON_RPC || 'https://rpc.ankr.com/polygon_mumbai',
  sei: env.SEI_RPC || 'https://sei-atlantic-rpc.allthatnode.com:1317',
  solana: env.SOLANA_RPC || 'https://api.devnet.solana.com', // https://api.testnet.solana.com (testnet = devnet for solana)
  sui: env.SUI_RPC || 'https://sui-testnet-rpc.allthatnode.com',
  terra: env.TERRA_RPC || 'https://bombay.stakesystems.io:2053',
  terra2: env.TERRA2_RPC || 'https://pisco-lcd.terra.dev',
  xpla: env.XPLA_RPC || 'https://dimension-rpc.xpla.dev',
} as const;

export const NETWORK_CONTRACTS = NETWORK === 'testnet' ? CONTRACTS.TESTNET : CONTRACTS.MAINNET;
export const NETWORK_RPCS_BY_CHAIN =
  NETWORK === 'testnet' ? RPCS_BY_CHAIN_TESTNET : RPCS_BY_CHAIN_MAINNET;

// Separating for now so if we max out infura we can keep Polygon going
export const POLYGON_ROOT_CHAIN_RPC = 'https://rpc.ankr.com/eth';
export const POLYGON_ROOT_CHAIN_ADDRESS = '0x86E4Dc95c7FBdBf52e33D563BbDB00823894C287';
// Optimism watcher relies on finalized calls which don't work right on Ankr
export const OPTIMISM_CTC_CHAIN_RPC = env.ETHEREUM_RPC;
export const OPTIMISM_CTC_CHAIN_ADDRESS = '0x5E4e65926BA27467555EB562121fac00D24E9dD2';

export const ALGORAND_INFO = {
  appid: Number(NETWORK_CONTRACTS.algorand.core),
  algodToken: '',
  algodServer: NETWORK_RPCS_BY_CHAIN.algorand,
  algodPort: 443,
  server: 'https://mainnet-idx.algonode.cloud',
  port: 443,
  token: '',
} as const;

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
