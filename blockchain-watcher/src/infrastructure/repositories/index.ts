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
export * from "./EvmJsonRPCBlockRepository";
export * from "./PromStatRepository";
export * from "./StaticJobRepository";
export * from "./Web3SolanaSlotRepository";
