import { CHAINS } from "@certusone/wormhole-sdk";

const chainIdToNameMap = Object.fromEntries(
  Object.entries(CHAINS).map(([key, value]) => [value, key])
);
const chainIdToName = (chainId: number) =>
  chainIdToNameMap[chainId] || "Unknown";
export default chainIdToName;
