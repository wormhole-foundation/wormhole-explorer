// Monkey patching BigInt serialization
if (!("toJSON" in BigInt.prototype)) {
  Object.defineProperty(BigInt.prototype, "toJSON", {
    get() {
      return () => String(this);
    },
  });
}

export * from "./jobs/metadata/FileMetadataRepository";
export * from "./jobs/metadata/PostgresMetadataRepository";
export * from "./jobs/execution/PostgresJobExecutionRepository";
export * from "./jobs/execution/InMemoryJobExecutionRepository";
export * from "./SnsEventRepository";
export * from "./evm/EvmJsonRPCBlockRepository";
export * from "./evm/BscEvmJsonRPCBlockRepository";
export * from "./evm/ArbitrumEvmJsonRPCBlockRepository";
export * from "./PromStatRepository";
export * from "./jobs/StaticJobRepository";
export * from "./solana/Web3SolanaSlotRepository";
export * from "./solana/RateLimitedSolanaSlotRepository";
