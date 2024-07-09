// Monkey patching BigInt serialization
if (!("toJSON" in BigInt.prototype)) {
  Object.defineProperty(BigInt.prototype, "toJSON", {
    get() {
      return () => String(this);
    },
  });
}

export * from "./FileMetadataRepository";
export * from "./SnsEventRepository";
export * from "./InfluxEventRepository";
export * from "./evm/EvmJsonRPCBlockRepository";
export * from "./evm/BscEvmJsonRPCBlockRepository";
export * from "./evm/ArbitrumEvmJsonRPCBlockRepository";
export * from "./evm/MoonbeamEvmJsonRPCBlockRepository";
export * from "./evm/PolygonEvmJsonRPCBlockRepository";
export * from "./PromStatRepository";
export * from "./StaticJobRepository";
export * from "./solana/Web3SolanaSlotRepository";
export * from "./solana/RateLimitedSolanaSlotRepository";
export * from "./sui/SuiJsonRPCBlockRepository";
export * from "./wormchain/WormchainJsonRPCBlockRepository";
export * from "./algorand/AlgorandJsonRPCBlockRepository";
