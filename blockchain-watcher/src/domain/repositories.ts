import { RunPollingJob } from "./actions/RunPollingJob";
import { EvmBlock, EvmLog, EvmLogFilter, Handler, JobDefinition } from "./entities";

export interface EvmBlockRepository {
  getBlockHeight(finality: string): Promise<bigint>;
  getBlocks(blockNumbers: Set<bigint>): Promise<Record<string, EvmBlock>>;
  getFilteredLogs(filter: EvmLogFilter): Promise<EvmLog[]>;
}

export interface MetadataRepository<Metadata> {
  get(id: string): Promise<Metadata | undefined>;
  save(id: string, metadata: Metadata): Promise<void>;
}

export interface StatRepository {
  count(id: string, labels: Record<string, any>): void;
  measure(id: string, value: bigint, labels: Record<string, any>): void;
  report: () => Promise<string>;
}

export interface JobRepository {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getSource(jobDef: JobDefinition): RunPollingJob;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
