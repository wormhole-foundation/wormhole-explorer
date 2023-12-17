import { RunPollingJob } from "./actions/RunPollingJob";
import { EvmBlock, EvmLog, EvmLogFilter, EvmTag, Handler, JobDefinition, solana } from "./entities";
import { ConfirmedSignatureInfo } from "./entities/solana";
import { Fallible, SolanaFailure } from "./errors";

export interface EvmBlockRepository {
  getBlockHeight(chain: string, finality: string): Promise<bigint>;
  getBlocks(chain: string, blockNumbers: Set<bigint>): Promise<Record<string, EvmBlock>>;
  getFilteredLogs(chain: string, filter: EvmLogFilter): Promise<EvmLog[]>;
  getTransactionReceipt(chain: string, hash: string): Promise<string>;
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
