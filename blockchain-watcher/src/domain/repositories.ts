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
