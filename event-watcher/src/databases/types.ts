import { ChainId, ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import JsonDB from './JsonDB';
import MongoDB from './MongoDB';

export type DBOptionTypes = MongoDB | JsonDB;
export interface DBImplementation {
  start(): Promise<void>;
  connect(): Promise<void>;
  getResumeBlockByChain(chain: ChainName): Promise<number | null>;
  getLastBlocksProcessed(): Promise<void>;
  getLastBlockByChain(chain: ChainName): string | null;
  storeWhTxs(chain: ChainName, whTxs: WHTransaction[]): Promise<void>;
  storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void>;
}

export type VaasByBlock = { [blockInfo: string]: string[] };

export type WHTransaction = {
  id: string;
  eventLog: EventLog;
  status: string;
};

export type EventLog = {
  emitterChain: number;
  emitterAddr: string;
  sequence: number;
  txHash: string;
  blockNumber: string | number;
  unsignedVaa: Buffer | Uint8Array | string;
  sender: string;
  indexedAt: Date | number | string;
  createdAt?: Date | number | string;
  updatedAt?: Date | number | string;
  revision?: number;
};

type LastBlockItem = {
  blockNumber: number;
  chainId: number;
  createdAt: Date | string;
  updatedAt: Date | string;
};

type LastBlockByChainWithId = LastBlockItem & {
  id: string;
};

type LastBlockByChainWith_Id = LastBlockItem & {
  _id: string;
};

export type LastBlockByChain = LastBlockByChainWith_Id | LastBlockByChainWithId;
