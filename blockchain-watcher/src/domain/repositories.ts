import { IbcTransaction, WormchainBlockLogs, CosmosRedeem } from "./entities/wormchain";
import { AptosEvent, AptosTransaction } from "./entities/aptos";
import { SuiTransactionBlockReceipt } from "./entities/sui";
import { Fallible, SolanaFailure } from "./errors";
import { ConfirmedSignatureInfo } from "./entities/solana";
import { RunDeploymentContract } from "./actions/RunDeploymentContract";
import { AlgorandTransaction } from "./entities/algorand";
import { TransactionFilter } from "./actions/aptos/PollAptos";
import { RunRPCHealthcheck } from "./actions/RunRPCHealthcheck";
import { CosmosTransaction } from "./entities/cosmos";
import { NearTransaction } from "./entities/near";
import { RunPollingJob } from "./actions";
import { Filter } from "./actions/cosmos/types";
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
  healthCheck(chain: string, finality: EvmTag, cursor: bigint): Promise<ProviderHealthCheck[]>;
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
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
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
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
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
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
}

export interface WormchainRepository {
  getBlockHeight(chain: string): Promise<bigint | undefined>;
  getBlockLogs(
    chain: string,
    blockNumber: bigint,
    attributesTypes: string[]
  ): Promise<WormchainBlockLogs>;
  getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]>;
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
}

export interface CosmosRepository {
  getTransactions(
    filter: Filter,
    blockBatchSize: number,
    chain: string
  ): Promise<CosmosTransaction[]>;
  getBlockTimestamp(blockNumber: bigint, chain: string): Promise<number | undefined>;
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
}

export interface AlgorandRepository {
  getTransactions(
    applicationId: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<AlgorandTransaction[]>;
  getBlockHeight(): Promise<bigint | undefined>;
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
}

export interface NearRepository {
  getTransactions(contract: string, fromBlock: bigint, toBlock: bigint): Promise<NearTransaction[]>;
  getBlockHeight(commitment: string): Promise<bigint | undefined>;
  healthCheck(chain: string, finality: string, cursor: bigint): Promise<ProviderHealthCheck[]>;
}

export interface MetadataRepository<Metadata> {
  get(id: string): Promise<Metadata | undefined>;
  save(id: string, metadata: Metadata): Promise<void>;
}

export interface StatRepository {
  count(id: string, labels: Record<string, any>, increase?: number): void;
  measure(id: string, value: bigint | number, labels: Record<string, any>): void;
  report: () => Promise<string>;
}

export type ProviderHealthCheck = {
  isHealthy: boolean;
  latency?: number;
  height: bigint | undefined;
  url: string;
};

export interface JobRepository {
  getDeploymentContract(jobDef: JobDefinition): RunDeploymentContract;
  getJobDefinitions(): Promise<JobDefinition[]>;
  getRPCHealthcheck(jobsDef: JobDefinition[]): RunRPCHealthcheck;
  getPollingJob(jobDef: JobDefinition): RunPollingJob;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
