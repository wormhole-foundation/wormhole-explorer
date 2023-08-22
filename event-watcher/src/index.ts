import dotenv from 'dotenv';
dotenv.config();

import { getDB } from './databases/utils';
import { getSNS } from './services/SNS/utils';
import { makeFinalizedWatcher } from './watchers/utils';
import { InfrastructureController } from './infrastructure/infrastructure.controller';
import { createServer } from './builder/server';
import { env, evmChains } from './config';
import { DBOptionTypes } from './databases/types';
import { SNSOptionTypes } from './services/SNS/types';
class EventWatcher {
  private infrastructureController = new InfrastructureController();

  constructor(private db: DBOptionTypes, private sns: SNSOptionTypes) {
    this.setup();
  }

  async setup() {
    await this.startServer();
  }

  async startServer() {
    const port = Number(env.PORT) || 3005;
    const server = await createServer(port);

    server.get('/ready', { logLevel: 'silent' }, this.infrastructureController.ready);
    server.get('/health', { logLevel: 'silent' }, this.infrastructureController.health);

    server.listen({ port, host: '0.0.0.0' }, (err: any, address: any) => {
      if (err) process.exit(1);
      console.log(`Server listening at ${address}`);
    });
  }

  async run() {
    await this.db.start();

    // for (const chain of supportedChains) {
    for (const chain of evmChains) {
      const watcher = makeFinalizedWatcher(chain);
      watcher.setDB(this.db);
      watcher.setServices(this.sns);
      watcher.watch();
    }
  }
}

// Init and run the event watcher
const db: DBOptionTypes = getDB();
const sns: SNSOptionTypes = getSNS();
const eventWatcher = new EventWatcher(db, sns);
eventWatcher.run();
