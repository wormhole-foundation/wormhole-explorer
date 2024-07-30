// Source chain events ABIs
export const WORMHOLE_SEND_TRANSCEIVER_MESSAGE_ABI = [
  {
    anonymous: false,
    inputs: [
      { indexed: false, internalType: "uint16", name: "recipientChain", type: "uint16" },
      {
        components: [
          { internalType: "bytes32", name: "sourceNttManagerAddress", type: "bytes32" },
          { internalType: "bytes32", name: "recipientNttManagerAddress", type: "bytes32" },
          { internalType: "bytes", name: "nttManagerPayload", type: "bytes" },
          { internalType: "bytes", name: "transceiverPayload", type: "bytes" },
        ],
        indexed: false,
        internalType: "struct TransceiverStructs.TransceiverMessage",
        name: "message",
        type: "tuple",
      },
    ],
    name: "SendTransceiverMessage",
    type: "event",
  },
];

// abi ref: https://sepolia.etherscan.io/address/0xcc6e5c994de73e8a115263b1b512e29b2026df55#code
export const AXELAR_SEND_TRANSCEIVER_MESSAGE_ABI = [
  {
    anonymous: false,
    inputs: [
      { indexed: true, internalType: "uint16", name: "recipientChainId", type: "uint16" },
      { indexed: false, internalType: "bytes", name: "nttManagerMessage", type: "bytes" },
      {
        indexed: true,
        internalType: "bytes32",
        name: "recipientNttManagerAddress",
        type: "bytes32",
      },
      { indexed: true, internalType: "bytes32", name: "refundAddress", type: "bytes32" },
    ],
    name: "SendTransceiverMessage",
    type: "event",
  },
];

// Target chain events ABIs
export const TRANSFER_REDEEMED_ABI = ["event TransferRedeemed(bytes32 indexed digest);"];
export const RECEIVED_RELAYED_MESSAGE_ABI = [
  "event ReceivedRelayedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress)",
];
export const RECEIVED_MESSAGE_ABI = [
  "event ReceivedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress, uint64 sequence)",
];
export const MESSAGE_ATTESTED_TO_ABI = [
  "event MessageAttestedTo (bytes32 digest, address transceiver, uint8 index)",
];
