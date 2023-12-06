import { EVMWatcher } from './EVMWatcher';

export class BSCWatcher extends EVMWatcher { s
  constructor() {
    super('bsc');
  }

  override async getFinalizedBlockNumber(): Promise<number> {
    const latestBlock = await super.getFinalizedBlockNumber();
    return Math.max(latestBlock - 15, 0);
  }
}
