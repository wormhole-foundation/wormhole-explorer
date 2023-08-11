import { EVMWatcher } from './EVMWatcher';

export class BSCWatcher extends EVMWatcher {
  constructor() {
    super('bsc');
  }
  async getFinalizedBlockNumber(): Promise<number> {
    const latestBlock = await super.getFinalizedBlockNumber();
    return Math.max(latestBlock - 15, 0);
  }
}
