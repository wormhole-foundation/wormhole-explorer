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
  txHash: string | null;
  sender: string | null;
  payload: any;
  blockNumber: number | string | null;
  indexedAt?: string | number;
  updatedAt?: string | number;
  createdAt?: string | number;
}

export type LastBlockByChain = { [chain in ChainId]?: string };
