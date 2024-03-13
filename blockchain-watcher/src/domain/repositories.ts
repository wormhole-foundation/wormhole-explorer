import {
  Checkpoint,
  SuiEventFilter,
  TransactionFilter as SuiTransactionFilter,
} from "@mysten/sui.js/client";
import { RunPollingJob } from "./actions/RunPollingJob";
import {
  EvmBlock,
  EvmLog,
  EvmLogFilter,
  EvmTag,
  Handler,
  JobDefinition,
  Range,
  ReceiptTransaction,
  solana,
} from "./entities";
import { ConfirmedSignatureInfo } from "./entities/solana";
import { Fallible, SolanaFailure } from "./errors";
import { SuiTransactionBlockReceipt } from "./entities/sui";
import { TransactionFilter } from "./actions/aptos/PollAptos";
import { AptosEvent, AptosTransaction } from "./entities/aptos";

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
  getTransactionsByVersion(
    events: AptosEvent[] | AptosTransaction[],
    filter: TransactionFilter
  ): Promise<AptosTransaction[]>;
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
