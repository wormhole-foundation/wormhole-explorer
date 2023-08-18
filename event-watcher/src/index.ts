import * as dotenv from 'dotenv';
dotenv.config();

import { ChainName, EVMChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { initDb } from './databases/utils';
import { makeFinalizedWatcher } from './watchers/utils';
import { InfrastructureController } from './infrastructure/infrastructure.controller';
import { createServer } from './builder/server';

// EVM Chains not supported
// aurora, gnosis, neon, sepolia

const evmChains: EVMChainName[] = [
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

const supportedChains: ChainName[] = [
  ...evmChains,
  'algorand',
  'aptos',
  'injective',
  'near',
  'solana',
  'sui',
  'terra',
  'terra2',
  'xpla',
];

const db = initDb();
const infrastructureController = new InfrastructureController();

const startServer = async () => {
  const port = Number(process.env.PORT) || 3005;
  const server = await createServer(port);

  server.get('/ready', { logLevel: 'silent' }, infrastructureController.ready);
  server.get('/health', { logLevel: 'silent' }, infrastructureController.health);

  server.listen({ port, host: '0.0.0.0' }, (err: any, address: any) => {
    if (err) {
      process.exit(1);
    }
    console.log(`Server listening at ${address}`);
  });
};

startServer();

const start = async () => {
  // We wait to the database to fetch the `latestBlocks` (avoid multi requests)
  // Im trying not to change too much the codebase.
  await db.getLastBlockByChain('unset');

  // for (const chain of supportedChains) {
  for (const chain of evmChains) {
    makeFinalizedWatcher(chain).watch();
  }
};

start();
