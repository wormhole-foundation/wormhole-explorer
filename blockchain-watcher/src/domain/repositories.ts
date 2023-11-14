import { EvmBlock, EvmLog, EvmLogFilter } from "./entities";

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
