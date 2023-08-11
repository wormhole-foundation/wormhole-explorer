import { ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { getLogger, WormholeLogger } from '../utils/logger';
import { VaasByBlock } from './types';

export class Database {
  logger: WormholeLogger;
  constructor() {
    this.logger = getLogger('db');
  }
  static filterEmptyBlocks(vaasByBlock: VaasByBlock): VaasByBlock {
    const filteredVaasByBlock: VaasByBlock = {};
    for (const [block, vaas] of Object.entries(vaasByBlock)) {
      if (vaas.length > 0) filteredVaasByBlock[block] = [...vaas];
    }
    return filteredVaasByBlock;
  }
  async getLastBlockByChain(chain: ChainName): Promise<string | null> {
    throw new Error('Not Implemented');
  }
  async storeVaasByBlock(chain: ChainName, vaasByBlock: VaasByBlock): Promise<void> {
    throw new Error('Not Implemented');
  }
}
