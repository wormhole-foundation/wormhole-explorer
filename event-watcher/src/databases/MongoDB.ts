import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import BaseDB from './BaseDB';
import { VaaLog } from './types';
import * as mongoDB from 'mongodb';
import { env } from '../config';

const WORMHOLE_TX_COLLECTION: string = 'wormholeTx';
const WORMHOLE_LAST_BLOCK_COLLECTION: string = 'lastBlockByChain';

export default class MongoDB extends BaseDB {
  private client: mongoDB.MongoClient | null = null;
  private db: mongoDB.Db | null = null;
  private wormholeTxCollection: mongoDB.Collection | null = null;
  private lastTxBlockByChainCollection: mongoDB.Collection | null = null;

  constructor() {
    super('MongoDB');
    console.log('[MongoDB]', 'Connecting...');
    this.client = new mongoDB.MongoClient(env.MONGODB_URI as string);
    this.db = this.client.db(env.MONGODB_DATABASE ?? 'wormhole');
    this.wormholeTxCollection = this.db.collection(WORMHOLE_TX_COLLECTION);
    this.lastTxBlockByChainCollection = this.db.collection(WORMHOLE_LAST_BLOCK_COLLECTION);
  }

  async connect(): Promise<void> {
    try {
      await this.client?.connect();

      console.log('[MongoDB]', 'Connected');
    } catch (e) {
      throw new Error(`[MongoDB] Error: ${e}`);
    }
  }

  async disconnect(): Promise<void> {
    console.log('[MongoDB]', 'Disconnecting...');
    await this.client?.close();
    console.log('[MongoDB]', 'Disconnected');
  }

  async isConnected() {
    try {
      await this.db?.command({ ping: 1 });
      return true;
    } catch (error: unknown) {
      return false;
    }
  }

  async getLastBlocksProcessed(): Promise<void> {
    try {
      const latestBlocks = await this.lastTxBlockByChainCollection?.findOne({});
      const json = JSON.parse(JSON.stringify(latestBlocks));
      this.lastBlockByChain = json || {};
    } catch (error: unknown) {
      this.logger.warn(`Error while getting last blocks processed: ${error}`);
      this.lastBlockByChain = {};
    }
  }

  override async storeVaaLogs(_: ChainName, vaaLogs: VaaLog[]): Promise<void> {
    const adaptedVaaLogs = vaaLogs.map((vaaLog) => {
      const { id, payloadBuffer, ...rest } = vaaLog;
      return {
        ...rest,
        _id: id,
        vaa: payloadBuffer,
      };
    });

    try {
      // @ts-ignore - I want to pass a custom _id field, but TypeScript doesn't like it (ObjectId error)
      const response = await this.wormholeTxCollection?.insertMany(adaptedVaaLogs, {
        ordered: false,
      });

      if (response) {
        const { insertedIds } = response;
        Object.values(insertedIds).forEach((id) => {
          const vaaLog: VaaLog | undefined = vaaLogs?.find((vaaLog) => vaaLog.id === id.toString());
          if (vaaLog) {
            const { blockNumber, chainName, id, txHash, chainId } = vaaLog;
            this.logger.info({
              blockNumber,
              chainName,
              id,
              txHash,
              chainId,
              message: 'Save VAA log to MongoDB',
            });
          }
        });
      }
    } catch (e: unknown) {
      this.logger.error(`Error while storing VAA logs: ${e}`);
    }
  }

  override async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);

    try {
      await this.lastTxBlockByChainCollection?.findOneAndUpdate(
        {},
        {
          $set: {
            [chainId]: lastBlock,
            updatedAt: new Date().getTime(),
          },
        },
        {
          upsert: true,
        },
      );
    } catch (e: unknown) {
      this.logger.error(`Error while storing latest processed block: ${e}`);
    }
  }
}
