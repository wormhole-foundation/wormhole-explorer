import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { readFileSync, writeFileSync } from 'fs';
import { DB_LAST_BLOCK_FILE, JSON_DB_FILE } from '../consts';
import { Database } from './Database';
import { DB, LastBlockByChain, VaasByBlock } from './types';
import * as mongoDB from "mongodb";
import { SequenceNumber } from '@mysten/sui.js';

export const collections: { wormholeTx?: mongoDB.Collection } = {}

const ENCODING = 'utf8';
export class MongoDatabase extends Database {
  lastBlockByChain: LastBlockByChain;
  dbFile: string;
  dbLastBlockFile: string;
  client: mongoDB.MongoClient;
  db: mongoDB.Db;
  wormholeTx: mongoDB.Collection;
  constructor() {
    super();
    this.client = new mongoDB.MongoClient("mongodb://localhost:27017");
    this.client.connect();
    this.db  = this.client.db("wormhole");
    this.wormholeTx = this.db.collection("wormholeTx");
 
   // collections.games = gamesCollection;
       
    console.log(`Successfully connected to database: ${this.db.databaseName} `);
    //this.db = client.db("wormhole");
    this.lastBlockByChain = {};

    if (!process.env.DB_LAST_BLOCK_FILE) {
      this.logger.info(`no db file set, using default path=${DB_LAST_BLOCK_FILE}`);
    }
    this.dbFile = JSON_DB_FILE;
    this.dbLastBlockFile = DB_LAST_BLOCK_FILE;

    try {
      const rawLast = readFileSync(this.dbLastBlockFile, ENCODING);
      this.lastBlockByChain = JSON.parse(rawLast);
    } catch (e) {
      this.logger.warn('Failed to load DB, initiating a fresh one.');
    }
  }

  async getLastBlockByChain(chain: ChainName): Promise<string | null> {
    const chainId = coalesceChainId(chain);
    const blockInfo = this.lastBlockByChain[chainId];
    if (blockInfo) {
      const tokens = blockInfo.split('/');
      return chain === 'aptos' ? tokens.at(-1)! : tokens[0];
    }
    return null;
  }
  async storeVaasByBlock(chain: ChainName, vaasByBlock: VaasByBlock): Promise<void> {
    const chainId = coalesceChainId(chain);
    const filteredVaasByBlock = Database.filterEmptyBlocks(vaasByBlock);
    if (Object.keys(filteredVaasByBlock).length) {
    }

    // this will always overwrite the "last" block, so take caution if manually backfilling gaps
    const blockKeys = Object.keys(vaasByBlock).sort(
      (bk1, bk2) => Number(bk1.split('/')[0]) - Number(bk2.split('/')[0])
    );
    if (blockKeys.length) {
      this.lastBlockByChain[chainId] = blockKeys[blockKeys.length - 1];
      this.wormholeTx.insertOne({chainId: chainId, block: this.lastBlockByChain[chainId], data: vaasByBlock});

      //writeFileSync(this.dbLastBlockFile, JSON.stringify(this.lastBlockByChain), ENCODING);
    }
  }

  async storeVaa(chain: ChainName, txHash: string, vaa_id:string, payload: string): Promise<void> {
    const chainId = coalesceChainId(chain);
    this.wormholeTx.insertOne({chainId: chainId, txHash: txHash, vaa_id: vaa_id, payload: payload});
  }


}

