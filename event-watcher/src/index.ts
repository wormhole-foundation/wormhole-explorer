import dotenv from 'dotenv';
dotenv.config();

import { getDB } from './databases/utils';
import { getSNS } from './services/SNS/utils';
import { makeFinalizedWatcher } from './watchers/utils';
import { InfrastructureController } from './infrastructure/infrastructure.controller';

import { supportedChains } from './config';
import { DBOptionTypes } from './databases/types';
import { SNSOptionTypes } from './services/SNS/types';
import WebServer from './services/WebServer';

class EventWatcher {
  constructor(private db: DBOptionTypes, private sns: SNSOptionTypes) {
    console.log('[EventWatcher]', 'Initializing...');
  }

  async run() {
    await this.db.start();

    // for (const chain of supportedChains) {
    //   const watcher = makeFinalizedWatcher(chain);
    //   watcher.setDB(this.db);
    //   watcher.setServices(this.sns);
    //   watcher.watch();
    // }

    // TEST
    const watcher = makeFinalizedWatcher('ethereum');
    watcher.setDB(this.db);
    watcher.setServices(this.sns);
    watcher.watch();
  }
}

(async () => {
  console.log('[APP]', 'Initializing...');
  console.log('--- --- --- --- ---');

  // Init and run the web server
  const infrastructureController = new InfrastructureController();
  const webServer = new WebServer(infrastructureController);
  await webServer.start();

  console.log('--- --- --- --- ---');

  // Init and run the event watcher
  const db: DBOptionTypes = getDB();
  const sns: SNSOptionTypes = getSNS();
  const eventWatcher = new EventWatcher(db, sns);
  await eventWatcher.run();
})();
