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
import { WatcherOptionTypes } from './watchers/types';
import { ChainName } from '@certusone/wormhole-sdk';

class EventWatcher {
  private watchers: WatcherOptionTypes[] = [];

  constructor(private db: DBOptionTypes, private sns: SNSOptionTypes) {
    console.log('[EventWatcher]', 'Initializing...');
  }

  async run() {
    await this.db.start();

    // for (const chain of supportedChains) {
    //   try {
    //     const watcher = makeFinalizedWatcher(chain);
    //     this.watchers.push(watcher);
    //     watcher.setDB(this.db);
    //     watcher.setServices(this.sns);
    //     watcher.watch();
    //   } catch (error: unknown) {
    //     console.warn(error);
    //   }
    // }

    // TEST
    {
      const chainName = 'ethereum22';
      try {
        const watcher = makeFinalizedWatcher(chainName as ChainName);
        this.watchers.push(watcher);
        watcher.setDB(this.db);
        watcher.setServices(this.sns);
        watcher.watch();
      } catch (error: unknown) {
        console.warn(error);
      }
    }
  }

  async stop() {
    for (const watcher of this.watchers) {
      await watcher.stop();
      console.log(`[${watcher.chain}] Watcher Stopped`);
    }
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

  // Handle shutdown
  const handleShutdown = async () => {
    console.log('--- --- --- --- ---');
    console.log('[APP]', 'Shutting down...');
    try {
      await eventWatcher.stop();
      await webServer.stop();
      await db.stop();

      console.log('[APP]', 'Exited as code 0');
      process.exit();
    } catch (error: unknown) {
      console.log('[APP]', 'Exited as code 1');
      process.exit(1);
    }
  };

  process.on('SIGINT', handleShutdown);
  process.on('SIGTERM', handleShutdown);
})();
