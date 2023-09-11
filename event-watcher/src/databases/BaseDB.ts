import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../common/consts';
import { getLogger, WormholeLogger } from '../utils/logger';
import { DBImplementation, LastBlockByChain, WHTransaction } from './types';
abstract class BaseDB implements DBImplementation {
  public logger: WormholeLogger;
  public lastBlocksByChain: LastBlockByChain[] = [];

  constructor(private readonly dbTypeName: string = '') {
    this.logger = getLogger(dbTypeName || 'db');
    this.lastBlocksByChain = [];
    this.logger.info(`Initializing as ${this.dbTypeName}...`);
  }

  public async start(): Promise<void> {
    this.logger.info('Starting...');

    await this.connect();
    await this.getLastBlocksProcessed();
    this.logger.info(`Connected as ${this.dbTypeName}`);
  }

  public async stop(): Promise<void> {
    this.logger.info('Stopping...');

    await this.disconnect();
  }

  public async getResumeBlockByChain(chain: ChainName): Promise<number | null> {
    const lastBlock = this.getLastBlockByChain(chain);
    const initialBlock = INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN[chain];

    if (lastBlock) return Number(lastBlock) + 1;
    if (initialBlock) return Number(initialBlock);
    return null;
  }

  public getLastBlockByChain(chain: ChainName): string | null {
    const item = this.lastBlocksByChain.find((item) => {
      if ('_id' in item) return item._id === chain;
      if ('id' in item) return item.id === chain;
      return false;
    });

    if (item) {
      const blockNumber = item.blockNumber;

      if (blockNumber) {
        const tokens = String(blockNumber)?.split('/');
        return chain === 'aptos' ? tokens.at(-1)! : tokens[0];
      }
    }

    return null;
  }

  abstract connect(): Promise<void>;
  abstract disconnect(): Promise<void>;
  abstract isConnected(): Promise<boolean>;
  abstract getLastBlocksProcessed(): Promise<void>;
  abstract storeWhTxs(chain: ChainName, whTxs: WHTransaction[]): Promise<void>;
  abstract storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void>;
}

export default BaseDB;
