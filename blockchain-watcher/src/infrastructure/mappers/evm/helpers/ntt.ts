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

const transceiverMessageLayout = <
  const MP extends CustomizableBytes = undefined,
  const TP extends CustomizableBytes = undefined
>(
  prefix: Prefix,
  nttManagerPayload?: MP,
  transceiverPayload?: TP
) =>
  [
    prefixItem(prefix),
    { name: "sourceNttManager", ...layoutItems.universalAddressItem },
    { name: "recipientNttManager", ...layoutItems.universalAddressItem },
    customizableBytes({ name: "nttManagerPayload", lengthSize: 2 }, nttManagerPayload),
    customizableBytes({ name: "transceiverPayload", lengthSize: 2 }, transceiverPayload),
  ] as const satisfies Layout;

export type TransceiverMessage<
  MP extends CustomizableBytes = undefined,
  TP extends CustomizableBytes = undefined
> = LayoutToType<ReturnType<typeof transceiverMessageLayout<MP, TP>>>;

export type WormholeTransceiverMessage<MP extends CustomizableBytes = undefined> = LayoutToType<
  ReturnType<typeof wormholeTransceiverMessageLayout<MP>>
>;

const wormholeTransceiverMessageLayout = <MP extends CustomizableBytes = undefined>(
  nttManagerPayload?: MP
) => transceiverMessageLayout([0x99, 0x45, 0xff, 0x10], nttManagerPayload, new Uint8Array(0));

export const deserializeWormholeTransceiverMessage = (message: Uint8Array) => {
  return deserializeLayout(
    wormholeTransceiverMessageLayout(nttManagerMessageLayout(nativeTokenTransferLayout)),
    message
  );
};
