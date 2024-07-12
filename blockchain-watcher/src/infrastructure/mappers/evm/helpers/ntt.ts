import { layoutItems } from "@wormhole-foundation/sdk-definitions";
import {
  CustomizableBytes,
  Layout,
  LayoutToType,
  customizableBytes,
  deserializeLayout,
} from "@wormhole-foundation/sdk-base";

export const deserializeNttMessageDigest = (digest: Uint8Array): Message => {
  return deserializeLayout(nttManagerMessageLayout(nativeTokenTransferLayout), digest);
};

export type Message = NttManagerMessage<typeof nativeTokenTransferLayout>;

export type NttManagerMessage<P extends CustomizableBytes = undefined> = LayoutToType<
  ReturnType<typeof nttManagerMessageLayout<P>>
>;

export const nttManagerMessageLayout = <const P extends CustomizableBytes = undefined>(
  customPayload?: P
) =>
  [
    { name: "id", binary: "bytes", size: 32 },
    { name: "sender", ...layoutItems.universalAddressItem },
    customizableBytes({ name: "payload", lengthSize: 2 }, customPayload),
  ] as const satisfies Layout;

type Prefix = readonly [number, number, number, number];

const prefixItem = (prefix: Prefix) =>
  ({
    name: "prefix",
    binary: "bytes",
    custom: Uint8Array.from(prefix),
    omit: true,
  } as const);

const trimmedAmountLayout = [
  { name: "decimals", binary: "uint", size: 1 },
  { name: "amount", binary: "uint", size: 8 },
] as const satisfies Layout;

const trimmedAmountItem = {
  binary: "bytes",
  layout: trimmedAmountLayout,
} as const;

/** Describes binary layout for a native token transfer payload */
export const nativeTokenTransferLayout = [
  prefixItem([0x99, 0x4e, 0x54, 0x54]),
  { name: "trimmedAmount", ...trimmedAmountItem },
  { name: "sourceToken", ...layoutItems.universalAddressItem },
  { name: "recipientAddress", ...layoutItems.universalAddressItem },
  { name: "recipientChain", ...layoutItems.chainItem() },
] as const satisfies Layout;
