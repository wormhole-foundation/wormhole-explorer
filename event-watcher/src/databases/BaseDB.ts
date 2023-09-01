import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../common/consts';
import { getLogger, WormholeLogger } from '../utils/logger';
import { DBImplementation, LastBlockByChain, VaaLog } from './types';
abstract class BaseDB implements DBImplementation {
  public logger: WormholeLogger;
  public lastBlockByChain: LastBlockByChain = {};

  constructor(private readonly dbTypeName: string = '') {
    console.log('[Database]', `Initializing as ${this.dbTypeName}...`);

    this.logger = getLogger(dbTypeName || 'db');
    this.lastBlockByChain = {};
  }

  public async start(): Promise<void> {
    console.log('[Database]', 'Starting...');

    await this.connect();
    await this.getLastBlocksProcessed();
    console.log('[Database]', `Connected as ${this.dbTypeName}`);
  }

  public async stop(): Promise<void> {
    console.log('[Database]', 'Stopping...');

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
    const chainId = coalesceChainId(chain);
    const blockInfo = this.lastBlockByChain?.[chainId];

    if (blockInfo) {
      const tokens = String(blockInfo)?.split('/');
      return chain === 'aptos' ? tokens.at(-1)! : tokens[0];
    }

    return null;
  }

  abstract connect(): Promise<void>;
  abstract disconnect(): Promise<void>;
  abstract isConnected(): Promise<boolean>;
  abstract getLastBlocksProcessed(): Promise<void>;
  abstract storeVaaLogs(chain: ChainName, vaaLogs: VaaLog[]): Promise<void>;
  abstract storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void>;
}

export default BaseDB;
