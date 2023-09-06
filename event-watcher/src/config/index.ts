import { ChainName, EVMChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';

export const env = {
  NODE_ENV: process.env.NODE_ENV || 'development',

  LOG_DIR: process.env.LOG_DIR,
  LOG_LEVEL: process.env.LOG_LEVEL || 'info',

  DB_SOURCE: process.env.DB_SOURCE || 'local',
  JSON_DB_FILE: process.env.JSON_DB_FILE || './db.json',
  JSON_LAST_BLOCK_FILE: process.env.JSON_LAST_BLOCK_FILE || './lastBlockByChain.json',

  PORT: process.env.PORT,

  MONGODB_URI: process.env.MONGODB_URI,
  MONGODB_DATABASE: process.env.MONGODB_DATABASE,

  SNS_SOURCE: process.env.SNS_SOURCE,
  AWS_SNS_REGION: process.env.AWS_SNS_REGION,
  AWS_SNS_TOPIC_ARN: process.env.AWS_SNS_TOPIC_ARN,
  AWS_SNS_SUBJECT: process.env.AWS_SNS_SUBJECT,
  AWS_ACCESS_KEY_ID: process.env.AWS_ACCESS_KEY_ID,
  AWS_SECRET_ACCESS_KEY: process.env.AWS_SECRET_ACCESS_KEY,

  CHAINS: process.env.CHAINS,

  ACALA_RPC: process.env.ACALA_RPC,
  ALGORAND_RPC: process.env.ALGORAND_RPC,
  APTOS_RPC: process.env.APTOS_RPC,
  ARBITRUM_RPC: process.env.ARBITRUM_RPC,
  AVALANCHE_RPC: process.env.AVALANCHE_RPC,
  BASE_RPC: process.env.BASE_RPC,
  BSC_RPC: process.env.BSC_RPC,
  CELO_RPC: process.env.CELO_RPC,
  ETHEREUM_RPC: process.env.ETHEREUM_RPC,
  FANTOM_RPC: process.env.FANTOM_RPC,
  INJECTIVE_RPC: process.env.INJECTIVE_RPC,
  KARURA_RPC: process.env.KARURA_RPC,
  KLAYTN_RPC: process.env.KLAYTN_RPC,
  MOONBEAM_RPC: process.env.MOONBEAM_RPC,
  NEAR_RPC: process.env.NEAR_RPC,
  OASIS_RPC: process.env.OASIS_RPC,
  OPTIMISM_RPC: process.env.OPTIMISM_RPC,
  POLYGON_RPC: process.env.POLYGON_RPC,
  SEI_RPC: process.env.SEI_RPC,
  SOLANA_RPC: process.env.SOLANA_RPC,
  SUI_RPC: process.env.SUI_RPC,
  TERRA_RPC: process.env.TERRA_RPC,
  TERRA2_RPC: process.env.TERRA2_RPC,
  XPLA_RPC: process.env.XPLA_RPC,
} as const;

// EVM Chains not supported
// aurora, gnosis, neon, sepolia

export const evmChains: EVMChainName[] = [
  'acala',
  'arbitrum',
  'avalanche',
  'base',
  'bsc',
  'celo',
  'ethereum',
  'fantom',
  'karura',
  'klaytn',
  'moonbeam',
  'oasis',
  'optimism',
  'polygon',
];

export const supportedChains: ChainName[] = [
  ...evmChains,
  'algorand',
  'aptos',
  'injective',
  'near',
  'sei',
  'solana',
  'sui',
  'terra',
  'terra2',
  'xpla',
];
