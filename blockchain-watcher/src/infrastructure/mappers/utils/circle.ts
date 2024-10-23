import { circle } from "@wormhole-foundation/sdk-connect";

export const toCirceChain = (env: string, domain: number) => {
  // Remove this when the SDK is updated to accept Noble as a domain with 4 value
  // @wormhole-foundation/sdk-base/dist/cjs/constants/circle.d.ts
  const environment = env === "mainnet" ? "Mainnet" : "Testnet";
  if (domain === 4) return "Noble";
  if (domain === 8) return "Sui";
  return circle.toCircleChain(environment, domain);
};

export enum MessageProtocol {
  Wormhole = "wormhole",
  None = "",
}
