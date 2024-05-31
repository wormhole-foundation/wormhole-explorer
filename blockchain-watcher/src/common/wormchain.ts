const chains: Map<number, string> = new Map([
  [19, "injective"],
  [20, "osmosis"],
  [4001, "evmos"],
  [4002, "kujira"],
]);

export function mapChain(chainId: number) {
  return chains.get(chainId) || "wormchain";
}
