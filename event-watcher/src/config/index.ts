import { ChainName, EVMChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';

export const env = {
  NODE_ENV: process.env.NODE_ENV || 'development',

  LOG_DIR: process.env.LOG_DIR,
  LOG_LEVEL: process.env.LOG_LEVEL || 'info',

  ETH_RPC: process.env.ETH_RPC,

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
  // 'terra',
  'terra2',
  'xpla',
];
