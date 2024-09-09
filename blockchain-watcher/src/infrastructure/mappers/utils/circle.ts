import { circle } from "@wormhole-foundation/sdk-connect";

export const toCirceChain = (env: string, domain: number) => {
  // Remove this when the SDK is updated to accept Noble as a domain with 4 value
  // @wormhole-foundation/sdk-base/dist/cjs/constants/circle.d.ts
  const environment = env === "mainnet" ? "Mainnet" : "Testnet";
  return domain === 4 ? "Noble" : circle.toCircleChain(environment, domain);
};

export enum MessageProtocol {
  Wormhole = "wormhole",
  None = "",
}
