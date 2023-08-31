import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { readFileSync, writeFileSync } from 'fs';
import { env } from '../config';
import BaseDB from './BaseDB';
import { VaaLog } from './types';

const ENCODING = 'utf8';

export default class JsonDB extends BaseDB {
  db: VaaLog[] = [];
  dbFile: string;
  dbLastBlockFile: string;

  constructor() {
    super('JsonDB');
    this.db = [];
    this.lastBlockByChain = {};
    this.dbFile = env.JSON_DB_FILE;
    this.dbLastBlockFile = env.JSON_LAST_BLOCK_FILE;
    console.log('[JsonDB]', 'Connecting...');
  }

  async connect(): Promise<void> {
    try {
      const rawDb = readFileSync(this.dbFile, ENCODING);
      this.db = rawDb ? JSON.parse(rawDb) : [];
      console.log('[JsonDB]', `${this.dbFile} file ready`);
    } catch (e) {
      this.logger.warn(`${this.dbFile} file does not exists, creating new file`);
      this.db = [];
    }
  }

  async getLastBlocksProcessed(): Promise<void> {
    try {
      const rawLastBlockByChain = readFileSync(this.dbLastBlockFile, ENCODING);
      this.lastBlockByChain = rawLastBlockByChain ? JSON.parse(rawLastBlockByChain) : {};
      console.log('[JsonDB]', `${this.dbLastBlockFile} file ready`);
    } catch (e) {
      this.logger.warn(`${this.dbLastBlockFile} file does not exists, creating new file`);
      this.lastBlockByChain = {};
    }
  }

  override async storeVaaLogs(chain: ChainName, vaaLogs: VaaLog[]): Promise<void> {
    this.db = [...this.db, ...vaaLogs];
    try {
      writeFileSync(this.dbFile, JSON.stringify(this.db, null, 2), ENCODING);
    } catch (e: unknown) {
      this.logger.error(`Error while storing VAA logs: ${e}`);
    }
  }

  override async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);
    this.lastBlockByChain[chainId] = String(lastBlock);

    try {
      writeFileSync(this.dbLastBlockFile, JSON.stringify(this.lastBlockByChain, null, 2), ENCODING);
    } catch (e: unknown) {
      this.logger.error(`Error while storing latest processed block: ${e}`);
    }
  }
}
