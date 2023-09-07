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
  storeVaaLogs(chain: ChainName, vaaLogs: VaaLog[]): Promise<void>;
  storeLatestProcessBlock(chain: ChainName, lastBlock: number): Promise<void>;
}

export type VaasByBlock = { [blockInfo: string]: string[] };
export interface VaaLog {
  id: string;
  trackId: string;
  chainId: number;
  chainName: string;
  emitter: string;
  sequence: number | string;
  txHash: string;
  payload: string | null;
  payloadBuffer?: Uint8Array | Buffer | null;
  blockNumber: number | string | null;
  indexedAt?: Date | string | number;
  updatedAt?: Date | string | number;
  createdAt?: Date | string | number;
}

type LastBlockItem = {
  blockNumber: number;
  chainId: number;
  createdAt: Date | string;
  indexedAt: Date | string;
  updatedAt: Date | string;
};

type LastBlockByChainWithId = LastBlockItem & {
  id: string;
};

type LastBlockByChainWith_Id = LastBlockItem & {
  _id: string;
};

export type LastBlockByChain = LastBlockByChainWith_Id | LastBlockByChainWithId;
