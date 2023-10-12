import { ChainName } from '@certusone/wormhole-sdk';
import BaseDB from '../databases/BaseDB';
import { WHTransaction, VaasByBlock, WHTransferRedeemed } from '../databases/types';
import BaseSNS from '../services/SNS/BaseSNS';
import { WormholeLogger } from '../utils/logger';
import { AlgorandWatcher } from './AlgorandWatcher';
import { AptosWatcher } from './AptosWatcher';
import { BSCWatcher } from './BSCWatcher';
import { CosmwasmWatcher } from './CosmwasmWatcher';
import { EVMWatcher } from './EVMWatcher';
import { InjectiveExplorerWatcher } from './InjectiveExplorerWatcher';
import { NearWatcher } from './NearWatcher';
import { SolanaWatcher } from './SolanaWatcher';
import { SuiWatcher } from './SuiWatcher';
import { TerraExplorerWatcher } from './TerraExplorerWatcher';

export type WatcherOptionTypes =
  | SolanaWatcher
  | EVMWatcher
  | BSCWatcher
  | AlgorandWatcher
  | AptosWatcher
  | NearWatcher
  | InjectiveExplorerWatcher
  | TerraExplorerWatcher
  | CosmwasmWatcher
  | SuiWatcher;
export interface WatcherImplementation {
  chain: ChainName;
  logger: WormholeLogger;
  maximumBatchSize: number;
  sns?: BaseSNS | null;
  db?: BaseDB;
  getFinalizedBlockNumber(): Promise<number>;
  getMessagesForBlocks(fromBlock: number, toBlock: number): Promise<VaasByBlock>;
  getWhEvents(
    fromBlock: number,
    toBlock: number,
  ): Promise<{
    whTxs: WHTransaction[];
    redeemedTxs: WHTransferRedeemed[];
    lastSequenceNumber: number | null;
  }>;
  getWhTxs(fromBlock: number, toBlock: number): Promise<WHTransaction[]>;
  getRedeemedTxs(fromBlock: number, toBlock: number): Promise<WHTransferRedeemed[]>;
  getLastSequenceNumber(whTxs: WHTransaction[]): Promise<number | null>;
  isValidBlockKey(key: string): boolean;
  isValidVaaKey(key: string): boolean;
  watch(): Promise<void>;
}

export type makeVAAInput = {
  timestamp: Date | string | number;
  nonce: number;
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  consistencyLevel: number;
  payloadAsHex: string;
};
