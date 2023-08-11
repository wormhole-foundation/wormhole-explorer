import * as dotenv from 'dotenv';
dotenv.config();

import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { initDb } from './databases/utils';
import { makeFinalizedWatcher } from './watchers/utils';
import { InfrastructureController } from "./infrastructure/infrastructure.controller";
import { createServer } from "./builder/server";


initDb();

const infrastructureController = new InfrastructureController();

const startServer = async () => {
  const port = Number(process.env.PORT) || 3005;
  const server =  await createServer(port);

  server.get("/ready", { logLevel: "silent" }, infrastructureController.ready);
  server.get("/health",{ logLevel: "silent" }, infrastructureController.health);

  server.listen({ port, host: "0.0.0.0" }, (err: any, address: any) => {
    if (err) {
      process.exit(1);
    }
    console.log(`Server listening at ${address}`);
  });
}

startServer();

const supportedChains: ChainName[] = [
  'solana',
  'ethereum',
  'bsc',
  'polygon',
  'avalanche',
  'oasis',
  'algorand',
  'fantom',
  'karura',
  'acala',
  'klaytn',
  'celo',
  'moonbeam',
  'arbitrum',
  'optimism',
  'aptos',
  'near',
  'terra2',
  'terra',
  'xpla',
  'injective',
  'sui',
  'base',
];

for (const chain of supportedChains) {
  makeFinalizedWatcher(chain).watch();
}
