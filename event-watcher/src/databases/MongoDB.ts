import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import BaseDB from './BaseDB';
import { LastBlockByChain, WHTransaction } from './types';
import * as mongoDB from 'mongodb';
import { env } from '../config';

const WORMHOLE_TX_COLLECTION: string = 'wormholeTxs';
const WORMHOLE_LAST_BLOCK_COLLECTION: string = 'lastBlocksByChain';

export default class MongoDB extends BaseDB {
  private client: mongoDB.MongoClient | null = null;
  private db: mongoDB.Db | null = null;
  private wormholeTxCollection: mongoDB.Collection | null = null;
  private lastTxBlockByChainCollection: mongoDB.Collection | null = null;

  constructor() {
    super('MongoDB');
    this.logger.info('Connecting...');
    this.client = new mongoDB.MongoClient(env.MONGODB_URI as string);
    this.db = this.client.db(env.MONGODB_DATABASE ?? 'wormhole');
    this.wormholeTxCollection = this.db.collection(WORMHOLE_TX_COLLECTION);
    this.lastTxBlockByChainCollection = this.db.collection(WORMHOLE_LAST_BLOCK_COLLECTION);
  }

  async connect(): Promise<void> {
    try {
      await this.client?.connect();

      this.logger.info('Connected');
    } catch (e) {
      throw new Error(`[MongoDB] Error: ${e}`);
    }
  }

  async disconnect(): Promise<void> {
    this.logger.info('Disconnecting...');
    await this.client?.close();
    this.logger.info('Disconnected');
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
      const lastBlocksByChain = await this.lastTxBlockByChainCollection?.find().toArray();
      this.lastBlocksByChain = (lastBlocksByChain as unknown as LastBlockByChain[]) || [];
    } catch (error: unknown) {
      this.logger.warn(`Error while getting last blocks processed: ${error}`);
      this.lastBlocksByChain = [];
    }
  }

  override async storeWhTxs(chainName: ChainName, whTxs: WHTransaction[]): Promise<void> {
    try {
      for (let i = 0; i < whTxs.length; i++) {
        let _upsertedId = null;
        let message = 'Save VAA log to MongoDB';
        const currentWhTx = whTxs[i];
        const { id, ...rest } = currentWhTx;

        // @ts-ignore - I want to pass a custom _id field, but TypeScript doesn't like it (ObjectId error)
        const whTxDocument = await this.wormholeTxCollection?.findOne({ _id: id });

        if (whTxDocument) {
          const response = await this.wormholeTxCollection?.findOneAndUpdate(
            {
              // @ts-ignore - I want to pass a custom _id field, but TypeScript doesn't like it (ObjectId error)
              _id: id,
            },
            {
              $set: {
                'eventLog.updatedAt': new Date(),
              },
              $inc: {
                'eventLog.revision': 1,
              },
            },
          );

          _upsertedId = response?.upsertedId;
          message = 'Update VAA log to MongoDB';
        } else {
          const response = await this.wormholeTxCollection?.insertOne({
            ...rest,
            // @ts-ignore - I want to pass a custom _id field, but TypeScript doesn't like it (ObjectId error)
            _id: id,
          });

          _upsertedId = response?.insertedId;
        }

        if (currentWhTx) {
          const { id, eventLog } = currentWhTx;
          const { blockNumber, txHash, emitterChain } = eventLog;

          this.logger.info({
            id,
            blockNumber,
            chainName,
            txHash,
            emitterChain,
            message,
          });
        }
      }
    } catch (e: unknown) {
      this.logger.error(`Error while storing VAA logs: ${e}`);
    }
  }

  override async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);

    try {
      await this.lastTxBlockByChainCollection?.findOneAndUpdate(
        {
          // @ts-ignore - I want to pass a custom _id field, but TypeScript doesn't like it (ObjectId error)
          _id: chain,
        },
        {
          $setOnInsert: {
            chainId,
            createdAt: new Date(),
          },
          $set: {
            blockNumber: lastBlock,
            updatedAt: new Date(),
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
