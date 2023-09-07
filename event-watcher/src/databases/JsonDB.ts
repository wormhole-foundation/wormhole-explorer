import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { readFileSync, writeFileSync } from 'fs';
import { env } from '../config';
import BaseDB from './BaseDB';
import { VaaLog } from './types';

const ENCODING = 'utf8';

type VaaLogWithVaa = Omit<VaaLog & { vaa: VaaLog['payload'] }, 'payload'>;
export default class JsonDB extends BaseDB {
  db: VaaLogWithVaa[] = [];
  dbFile: string;
  dbLastBlockFile: string;

  constructor() {
    super('JsonDB');
    this.db = [];
    this.lastBlocksByChain = [];
    this.dbFile = env.JSON_DB_FILE;
    this.dbLastBlockFile = env.JSON_LAST_BLOCK_FILE;
    this.logger.info('Connecting...');
  }

  async connect(): Promise<void> {
    try {
      const rawDb = readFileSync(this.dbFile, ENCODING);
      this.db = rawDb ? JSON.parse(rawDb) : [];
      this.logger.info(`${this.dbFile} file ready`);
    } catch (e) {
      this.logger.warn(`${this.dbFile} file does not exists, creating new file`);
      this.db = [];
    }
  }

  async disconnect(): Promise<void> {
    this.logger.info('Disconnecting...');
    this.logger.info('Disconnected');
  }

  async isConnected() {
    return true;
  }

  async getLastBlocksProcessed(): Promise<void> {
    try {
      const lastBlocksByChain = readFileSync(this.dbLastBlockFile, ENCODING);
      this.lastBlocksByChain = lastBlocksByChain ? JSON.parse(lastBlocksByChain) : [];
      this.logger.info(`${this.dbLastBlockFile} file ready`);
    } catch (e) {
      this.logger.warn(`${this.dbLastBlockFile} file does not exists, creating new file`);
      this.lastBlocksByChain = [];
    }
  }

  override async storeVaaLogs(_: ChainName, vaaLogs: VaaLog[]): Promise<void> {
    const adaptedVaaLogs = vaaLogs.map((vaaLog) => {
      const { payload, ...rest } = vaaLog;
      return {
        ...rest,
        vaa: payload,
        payloadBuffer: null,
      };
    });

    this.db = [...this.db, ...adaptedVaaLogs];

    try {
      writeFileSync(this.dbFile, JSON.stringify(this.db, null, 2), ENCODING);
    } catch (e: unknown) {
      this.logger.error(`Error while storing VAA logs: ${e}`);
    }
  }

  override async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);
    const updatedLastBlocksByChain = [...this.lastBlocksByChain];
    const itemIndex = updatedLastBlocksByChain.findIndex((item) => {
      if ('id' in item) return item.id === chain;
      return false;
    });

    if (itemIndex >= 0) {
      updatedLastBlocksByChain[itemIndex] = {
        ...updatedLastBlocksByChain[itemIndex],
        blockNumber: lastBlock,
        updatedAt: new Date(),
      };
    } else {
      updatedLastBlocksByChain.push({
        id: chain,
        blockNumber: lastBlock,
        chainId,
        createdAt: new Date(),
        indexedAt: new Date(),
        updatedAt: new Date(),
      });
    }

    this.lastBlocksByChain = updatedLastBlocksByChain;

    try {
      writeFileSync(
        this.dbLastBlockFile,
        JSON.stringify(this.lastBlocksByChain, null, 2),
        ENCODING,
      );
    } catch (e: unknown) {
      this.logger.error(`Error while storing latest processed block: ${e}`);
    }
  }
}
