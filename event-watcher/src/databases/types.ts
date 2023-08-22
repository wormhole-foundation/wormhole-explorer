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

export interface VaaLog {
  vaaId: string;
  chainId: number;
  chainName: string;
  emitter: string;
  sequence: number;
  txHash: string;
  sender: string;
  payload: any;
  blockNumber: number;
  indexedAt?: string | number;
  updatedAt?: string | number;
  createdAt?: string | number;
}

export type LastBlockByChain = { [chain in ChainId]?: string };
