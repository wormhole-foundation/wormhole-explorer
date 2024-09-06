import type { Layout, LayoutToType } from "@wormhole-foundation/sdk-base";
import { deserializeLayout } from "@wormhole-foundation/sdk-base";
import { layoutItems } from "@wormhole-foundation/sdk-definitions";

const { amountItem, circleDomainItem, circleNonceItem, universalAddressItem } = layoutItems;

const messageVersionItem = { binary: "uint", size: 4, custom: 0, omit: true } as const;

// https://developers.circle.com/stablecoin/docs/cctp-technical-reference#message
const circleBurnMessageLayout = [
  // messageBodyVersion is:
  // * immutable: https://github.com/circlefin/evm-cctp-contracts/blob/adb2a382b09ea574f4d18d8af5b6706e8ed9b8f2/src/TokenMessenger.sol#L107
  // * 0: https://etherscan.io/address/0xbd3fa81b58ba92a82136038b25adec7066af3155#readContract
  { name: "messageBodyVersion", ...messageVersionItem },
  { name: "burnToken", ...universalAddressItem },
  { name: "mintRecipient", ...universalAddressItem },
  { name: "amount", ...amountItem },
  { name: "messageSender", ...universalAddressItem },
] as const satisfies Layout;

export const circleMessageLayout = [
  // version is:
  // * immutable: https://github.com/circlefin/evm-cctp-contracts/blob/adb2a382b09ea574f4d18d8af5b6706e8ed9b8f2/src/MessageTransmitter.sol#L75
  // * 0: https://etherscan.io/address/0x0a992d191deec32afe36203ad87d7d289a738f81#readContract
  { name: "version", ...messageVersionItem },
  { name: "sourceDomain", ...circleDomainItem },
  { name: "destinationDomain", ...circleDomainItem },
  { name: "nonce", ...circleNonceItem },
  { name: "sender", ...universalAddressItem },
  { name: "recipient", ...universalAddressItem },
  { name: "destinationCaller", ...universalAddressItem },
  { name: "payload", binary: "bytes" },
] as const satisfies Layout;

const TOKEN_MESSENGER_CONTRACTS = [
  // Mainnet
  "0x000000000000000000000000bd3fa81b58ba92a82136038b25adec7066af3155", // Ethereum
  "0x0000000000000000000000006b25532e1060ce10cc3b0a99e5683b91bfde6982", // Avalanche
  "0x0000000000000000000000002b4069517957735be00cee0fadae88a26365528f", // Optimism
  "0x00000000000000000000000019330d10d9cc8751218eaf51e8885d058642e08a", // Arbitrum
  "0x0000000000000000000000001682ae6375c4e4a97e4b583bc394c861a46d8962", // Base
  "0x0000000000000000000000009daf8c91aefae50b9c0e69629d3f6ca40ca3b3fe", // Polygon
  // Testnet
  "0x0000000000000000000000009f3b8679c73c2fef8b59b4f3444d4e156fb70aa5", // Arbitrum, Ethereum, Base, Optimism and Polygon
  "0x000000000000000000000000eb08f243e5d3fcff26a9e38ae5520a669f4019d0", // Avalanche
];

export type CircleMessage<T = any> = Omit<LayoutToType<typeof circleMessageLayout>, "payload"> & {
  payload: T;
};

export type CircleBurnMessage = LayoutToType<typeof circleBurnMessageLayout>;

export type CircleProtocol = "cctp" | "unknown";

export const deserializeCircleHeader = (message: Uint8Array): CircleMessage => {
  return deserializeLayout(circleMessageLayout, message);
};

export const deserializeCCTPPayload = (payload: Uint8Array): CircleBurnMessage => {
  return deserializeLayout(circleBurnMessageLayout, payload);
};

export const deserializeCircleMessage = (
  bytes: Uint8Array
): [CircleProtocol, CircleMessage<Uint8Array | CircleBurnMessage>] => {
  const header = deserializeCircleHeader(bytes);

  if (TOKEN_MESSENGER_CONTRACTS.includes(header.sender.toString())) {
    const payload = deserializeCCTPPayload(header.payload);
    return [
      "cctp",
      {
        ...header,
        payload,
      },
    ];
  }

  return ["unknown", header];
};
