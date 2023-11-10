// Monkey patching BigInt serialization
if (!("toJSON" in BigInt.prototype)) {
  Object.defineProperty(BigInt.prototype, "toJSON", {
    get() {
      return () => String(this);
    },
  });
}

export * from "./FileMetadataRepo";
export * from "./SnsEventRepository";
export * from "./EvmJsonRPCBlockRepository";
