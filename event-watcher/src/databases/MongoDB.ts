import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { Database } from './Database';
import { LastBlockByChain, VaaLog, VaasByBlock } from './types';
import * as mongoDB from 'mongodb';

export class MongoDatabase extends Database {
  client: mongoDB.MongoClient;
  db: mongoDB.Db;
  wormholeTx: mongoDB.Collection;
  lastTxBlockByChain: mongoDB.Collection;
  lastBlockByChain: LastBlockByChain | null;

  constructor() {
    super();

    this.lastBlockByChain = null;
    this.client = new mongoDB.MongoClient(process.env.MONGODB_URI ?? 'mongodb://localhost:27017');
    this.connectDB();
    this.db = this.client.db('wormhole');
    this.wormholeTx = this.db.collection('wormholeTx');
    this.lastTxBlockByChain = this.db.collection('lastBlockByChain');
  }

  async connectDB() {
    await this.client.connect();
    console.log(`Successfully connected to database: ${this.db.databaseName} `);
  }

  async getLastBlockByChainFromDB() {
    const latestBlocks = await this.lastTxBlockByChain.findOne({});
    const json = JSON.parse(JSON.stringify(latestBlocks));
    this.lastBlockByChain = json;
  }

  async getLastBlockByChain(chain: ChainName): Promise<string | null> {
    if (!this.lastBlockByChain) await this.getLastBlockByChainFromDB();

    const chainId = coalesceChainId(chain);
    const blockInfo: string | undefined = this.lastBlockByChain?.[chainId];

    if (blockInfo) {
      const tokens = String(blockInfo)?.split('/');
      return chain === 'aptos' ? tokens.at(-1)! : tokens[0];
    }

    return null;
  }

  async storeVaasByBlock(chain: ChainName, vaasByBlock: VaasByBlock): Promise<void> {
    // const chainId = coalesceChainId(chain);
    // const filteredVaasByBlock = Database.filterEmptyBlocks(vaasByBlock);
    // if (Object.keys(filteredVaasByBlock).length) {
    // }
    // this will always overwrite the "last" block, so take caution if manually backfilling gaps
    // const blockKeys = Object.keys(vaasByBlock).sort(
    //   (bk1, bk2) => Number(bk1.split('/')[0]) - Number(bk2.split('/')[0]),
    // );
    // if (blockKeys.length) {
    //   this.lastBlockByChain[chainId] = blockKeys[blockKeys.length - 1];
    //   await this.wormholeTx.insertOne({
    //     chainId: chainId,
    //     block: this.lastBlockByChain[chainId],
    //     data: vaasByBlock,
    //   });
    // }
  }

  async storeVaaLogs(chain: ChainName, vaaLogs: VaaLog[]): Promise<void> {
    await this.wormholeTx.insertMany(vaaLogs);
  }

  async storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void> {
    const chainId = coalesceChainId(chain);

    await this.lastTxBlockByChain.findOneAndUpdate(
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
  }

  async storeVaa(chain: ChainName, txHash: string, vaa_id: string, payload: string): Promise<void> {
    const chainId = coalesceChainId(chain);
    this.wormholeTx.insertOne({
      chainId: chainId,
      txHash: txHash,
      vaa_id: vaa_id,
      payload: payload,
    });
  }
}
