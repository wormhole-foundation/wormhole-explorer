import { EvmBlock, EvmLog, EvmLogFilter } from "./entities";

export interface EvmBlockRepository {
  getBlockHeight(finality: string): Promise<bigint>;
  getBlocks(blockNumbers: Set<bigint>): Promise<EvmBlock[]>;
  getFilteredLogs(filter: EvmLogFilter): Promise<EvmLog[]>;
}

export interface MetadataRepository<Metadata> {
  getMetadata(id: string): Promise<Metadata | undefined>;
}
