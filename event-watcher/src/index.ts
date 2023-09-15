import dotenv from 'dotenv';
dotenv.config();

import { getDB } from './databases/utils';
import { getSNS } from './services/SNS/utils';
import { makeFinalizedWatcher } from './watchers/utils';
import { InfrastructureController } from './infrastructure/infrastructure.controller';

import { env, supportedChains } from './config';
import { DBOptionTypes } from './databases/types';
import { SNSOptionTypes } from './services/SNS/types';
import WebServer from './services/WebServer';
import { WatcherOptionTypes } from './watchers/types';
import { logInfo } from './utils/logger';
import { ChainName } from '@certusone/wormhole-sdk';

const version = '1.0.1';
class EventWatcher {
  private watchers: WatcherOptionTypes[] = [];

  constructor(private db: DBOptionTypes, private sns: SNSOptionTypes) {
    logInfo({ labels: ['EventWatcher'], message: `Initializing...` });
  }

  async run() {
    await this.db.start();

    const chains = env.CHAINS ? env.CHAINS.split(',') : supportedChains;

    for (const chain of chains) {
      try {
        const watcher = makeFinalizedWatcher(chain as ChainName);
        this.watchers.push(watcher);
        watcher.setDB(this.db);
        watcher.setServices(this.sns);
        watcher.watch();
      } catch (error: unknown) {
        logInfo({ labels: ['EventWatcher'], message: `${error}` });
      }
    }
  }

  async stop() {
    for (const watcher of this.watchers) {
      watcher.stop();
      logInfo({ labels: [watcher.chain], message: 'Watcher Stopped' });
    }
  }
}

const start = async () => {
  logInfo({ labels: ['App'], message: `Initializing... - v${version}` });
  // Dependencies / Instances
  const db: DBOptionTypes = getDB();
  const sns: SNSOptionTypes = getSNS();

  // Init and run the web server
  const infrastructureController = new InfrastructureController(db);
  const webServer = new WebServer(infrastructureController);
  await webServer.start();

  // Init and run the event watcher
  const eventWatcher = new EventWatcher(db, sns);
  await eventWatcher.run();

  // Handle shutdown
  const handleShutdown = async () => {
    logInfo({ labels: ['App'], message: 'Shutting down...' });
    try {
      await Promise.allSettled([eventWatcher.stop(), webServer.stop(), db.stop()]);

      logInfo({ labels: ['App'], message: 'Exited with code 0' });
      process.exit();
    } catch (error: unknown) {
      logInfo({ labels: ['App'], message: 'Exited with code 1' });
      process.exit(1);
    }
  };

  process.on('SIGINT', handleShutdown);
  process.on('SIGTERM', handleShutdown);
};

start();
