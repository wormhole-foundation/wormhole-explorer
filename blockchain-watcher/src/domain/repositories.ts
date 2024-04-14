import { AptosEvent, AptosTransaction } from "./entities/aptos";
import { SuiTransactionBlockReceipt } from "./entities/sui";
import { Fallible, SolanaFailure } from "./errors";
import { ConfirmedSignatureInfo } from "./entities/solana";
import { WormchainBlockLogs } from "./entities/wormchain";
import { TransactionFilter } from "./actions/aptos/PollAptos";
import { RunPollingJob } from "./actions/RunPollingJob";
import {
  TransactionFilter as SuiTransactionFilter,
  SuiEventFilter,
  Checkpoint,
} from "@mysten/sui.js/client";
import {
  ReceiptTransaction,
  JobDefinition,
  EvmLogFilter,
  EvmBlock,
  Handler,
  solana,
  EvmLog,
  EvmTag,
  Range,
} from "./entities";

export interface EvmBlockRepository {
  getBlockHeight(chain: string, finality: string): Promise<bigint>;
  getBlocks(
    chain: string,
    blockNumbers: Set<bigint>,
    isTransactionsPresent: boolean
  ): Promise<Record<string, EvmBlock>>;
  getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]>;
  getTransactionReceipt(
    chain: string,
    hashNumbers: Set<string>
  ): Promise<Record<string, ReceiptTransaction>>;
  getBlock(
    chain: string,
    blockNumberOrTag: EvmTag | bigint,
    isTransactionsPresent: boolean
  ): Promise<EvmBlock>;
}

export interface SolanaSlotRepository {
  getLatestSlot(commitment: string): Promise<number>;
  getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, SolanaFailure>>;
  getSignaturesForAddress(
    address: string,
    beforeSig: string,
    afterSig: string,
    limit: number,
    finality?: string
  ): Promise<ConfirmedSignatureInfo[]>;
  getTransactions(sigs: ConfirmedSignatureInfo[], finality?: string): Promise<solana.Transaction[]>;
}

export interface SuiRepository {
  getLastCheckpointNumber(): Promise<bigint>;
  getCheckpoint(sequence: string | bigint | number): Promise<Checkpoint>;
  getLastCheckpoint(): Promise<Checkpoint>;
  getCheckpoints(range: Range): Promise<Checkpoint[]>;
  getTransactionBlockReceipts(digests: string[]): Promise<SuiTransactionBlockReceipt[]>;
  queryTransactions(
    filter?: SuiTransactionFilter,
    cursor?: string
  ): Promise<SuiTransactionBlockReceipt[]>;
  queryTransactionsByEvent(
    filter: SuiEventFilter,
    cursor?: string
  ): Promise<SuiTransactionBlockReceipt[]>;
}

export interface AptosRepository {
  getTransactions(
    range: { from?: number | undefined; limit?: number | undefined } | undefined
  ): Promise<AptosTransaction[]>;
  getEventsByEventHandle(
    range: { from?: number | undefined; limit?: number | undefined } | undefined,
    filter: TransactionFilter
  ): Promise<AptosEvent[]>;
  getTransactionsByVersion(records: AptosEvent[] | AptosTransaction[]): Promise<AptosTransaction[]>;
}

export interface WormchainRepository {
  getBlockHeight(): Promise<bigint | undefined>;
  getBlockLogs(chainId: number, blockNumber: bigint): Promise<WormchainBlockLogs>;
}

export interface MetadataRepository<Metadata> {
  get(id: string): Promise<Metadata | undefined>;
  save(id: string, metadata: Metadata): Promise<void>;
}

export interface StatRepository {
  count(id: string, labels: Record<string, any>, increase?: number): void;
  measure(id: string, value: bigint, labels: Record<string, any>): void;
  report: () => Promise<string>;
}

export interface JobRepository {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getSource(jobDef: JobDefinition): RunPollingJob;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
