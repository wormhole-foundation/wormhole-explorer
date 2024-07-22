import { keccak256 } from "@wormhole-foundation/sdk-definitions";
import { ChainId } from "@wormhole-foundation/sdk-base";
import BN from "bn.js";
import { TransactionFoundAttributes } from "../../../../domain/entities";

// Ref: https://github.com/wormhole-foundation/wormhole-dashboard/blob/72a312dc7e7141c49e315943e0cd7664fe63afd4/watcher/src/watchers/NTTPayloads.ts
export class TransceiverMessage<A> {
  static prefix: Buffer;
  sourceNttManager: Buffer;
  recipientNttManager: Buffer;
  ntt_managerPayload: NttManagerMessage<A>;
  transceiverPayload: Buffer;

  constructor(
    sourceNttManager: Buffer,
    recipientNttManager: Buffer,
    ntt_managerPayload: NttManagerMessage<A>,
    transceiverPayload: Buffer
  ) {
    this.sourceNttManager = sourceNttManager;
    this.recipientNttManager = recipientNttManager;
    this.ntt_managerPayload = ntt_managerPayload;
    this.transceiverPayload = transceiverPayload;
  }

  static deserialize<A>(
    data: Buffer,
    deserializer: (data: Buffer) => NttManagerMessage<A>
  ): TransceiverMessage<A> {
    if (this.prefix == undefined) {
      throw new Error("Unknown prefix.");
    }
    const prefix = data.subarray(0, 4);
    if (!prefix.equals(this.prefix)) {
      throw new Error("Invalid transceiver prefix");
    }
    const sourceNttManager = data.subarray(4, 36);
    const recipientNttManager = data.subarray(36, 68);
    const ntt_managerPayloadLen = data.readUInt16BE(68);
    const ntt_managerPayload = deserializer(data.subarray(70, 70 + ntt_managerPayloadLen));
    const transceiverPayloadLen = data.readUInt16BE(70 + ntt_managerPayloadLen);
    const transceiverPayload = data.subarray(
      72 + ntt_managerPayloadLen,
      72 + ntt_managerPayloadLen + transceiverPayloadLen
    );
    return new TransceiverMessage(
      sourceNttManager,
      recipientNttManager,
      ntt_managerPayload,
      transceiverPayload
    );
  }

  static serialize<A>(
    msg: TransceiverMessage<A>,
    serializer: (payload: NttManagerMessage<A>) => Buffer
  ): Buffer {
    const payload = serializer(msg.ntt_managerPayload);
    if (msg.sourceNttManager.length != 32) {
      throw new Error("sourceNttManager must be 32 bytes");
    }
    if (msg.recipientNttManager.length != 32) {
      throw new Error("recipientNttManager must be 32 bytes");
    }
    const payloadLen = new BN(payload.length).toBuffer("be", 2);
    const transceiverPayloadLen = new BN(msg.transceiverPayload.length).toBuffer("be", 2);
    const buffer = Buffer.concat([
      this.prefix,
      msg.sourceNttManager,
      msg.recipientNttManager,
      payloadLen,
      payload,
      transceiverPayloadLen,
      msg.transceiverPayload,
    ]);
    return buffer;
  }
}

export class NttManagerMessage<A> {
  id: Buffer;
  sender: Buffer;
  payload: A;

  constructor(id: Buffer, sender: Buffer, payload: A) {
    if (id.length != 32) {
      throw new Error("id must be 32 bytes");
    }
    if (sender.length != 32) {
      throw new Error("sender must be 32 bytes");
    }
    this.id = id;
    this.sender = sender;
    this.payload = payload;
  }

  // This is the deserializer for the NttManagerMessage struct.
  // It follows the platform independent wire format.
  static deserialize = <A>(
    data: Buffer,
    deserializer: (data: Buffer) => A
  ): NttManagerMessage<A> => {
    const id = data.subarray(0, 32);
    const sender = data.subarray(32, 64);
    const payloadLen = data.readUint16BE(64);
    const payload = deserializer(data.subarray(66, 66 + payloadLen));
    return new NttManagerMessage(id, sender, payload);
  };

  // This is a different deserialization method from the one above.
  // This follows the Solana account serialization format (Borsh encoding).
  // This is used to deserialize the transceiverMessage account data into ValidatedTransceiverMessage.
  // ref: https://github.com/wormhole-foundation/example-native-token-transfers/blob/main/solana/programs/example-native-token-transfers/src/messages.rs#L7
  static deserializeAccountFormat = <A>(
    data: Buffer,
    deserializer: (data: Buffer) => A
  ): NttManagerMessage<A> => {
    const id = data.subarray(0, 32);
    const sender = data.subarray(32, 64);
    const payload = deserializer(data.subarray(64));
    return new NttManagerMessage(id, sender, payload);
  };

  static serialize = <A>(msg: NttManagerMessage<A>, serializer: (payload: A) => Buffer): Buffer => {
    const payload = serializer(msg.payload);
    return Buffer.concat([msg.id, msg.sender, new BN(payload.length).toBuffer("be", 2), payload]);
  };
}

export class NativeTokenTransfer {
  static prefix = Buffer.from([0x99, 0x4e, 0x54, 0x54]);
  trimmedAmount: TrimmedAmount;
  sourceToken: Buffer;
  recipientAddress: Buffer;
  recipientChain: number;

  constructor(
    sourceToken: Buffer,
    amount: TrimmedAmount,
    recipientChain: number,
    recipientAddress: Buffer
  ) {
    this.trimmedAmount = amount;
    this.sourceToken = sourceToken;
    this.recipientAddress = recipientAddress;
    this.recipientChain = recipientChain;
  }

  // This is the deserializer for the NativeTokenTransfer struct.
  // It follows the platform independent wire format through which both EVM and Solana implementations can communicate.
  // On Solana, the serialization is done manually via `Writeable` trait.
  // ref: https://github.com/wormhole-foundation/example-native-token-transfers/blob/main/solana/modules/ntt-messages/src/ntt.rs#L62
  static deserialize = (data: Buffer): NativeTokenTransfer => {
    const prefix = data.subarray(0, 4);
    if (!prefix.equals(NativeTokenTransfer.prefix)) {
      throw new Error("Invalid NTT prefix");
    }
    const amount = TrimmedAmount.deserialize(data.subarray(4, 13));
    const sourceToken = data.subarray(13, 45);
    const recipientAddress = data.subarray(45, 77);
    const recipientChain = data.readUInt16BE(77);
    return new NativeTokenTransfer(sourceToken, amount, recipientChain, recipientAddress);
  };

  // This is a different deserialization method from the one above.
  // This follows the Solana account serialization format (Borsh encoding).
  // This is used to deserialize the NTT struct from transceiverMessage account data.
  // Notice that the account format has no prefix as it uses Anchor Serialization and Deserialization.
  // ref: https://github.com/wormhole-foundation/example-native-token-transfers/blob/main/solana/modules/ntt-messages/src/ntt.rs#L11
  static deserializeAccountFormat = (data: Buffer): NativeTokenTransfer => {
    const amount = TrimmedAmount.deserializeAccount(data.subarray(0, 9));
    const sourceToken = data.subarray(9, 41);
    const recipientChain = data.readUInt16LE(41);
    const recipientAddress = data.subarray(43, 75);
    return new NativeTokenTransfer(sourceToken, amount, recipientChain, recipientAddress);
  };

  static serialize = (msg: NativeTokenTransfer): Buffer => {
    const buffer = Buffer.concat([
      NativeTokenTransfer.prefix,
      TrimmedAmount.serialize(msg.trimmedAmount),
      msg.sourceToken,
      msg.recipientAddress,
    ]);
    const recipientChain = Buffer.alloc(2);
    recipientChain.writeUInt16BE(msg.recipientChain, 0);
    return Buffer.concat([buffer, recipientChain]);
  };
}

export class TrimmedAmount {
  amount: bigint;
  decimals: number;

  constructor(amount: bigint, decimals: number) {
    this.amount = amount;
    this.decimals = decimals;
  }

  // Similar to the NativeTokenTransfer, this deserializer follows the platform independent wire format.
  // writable ref: https://github.com/wormhole-foundation/example-native-token-transfers/blob/main/solana/modules/ntt-messages/src/trimmed_amount.rs#L107
  static deserialize(data: Buffer): TrimmedAmount {
    const decimals = data.readUInt8(0);
    const amount = data.readBigUInt64BE(1);
    return new TrimmedAmount(amount, decimals);
  }

  // This is a different deserialization method from the one above.
  // This follows the Solana account serialization format (Borsh encoding).
  // This is used to deserialize the TrimmedAmount struct from NativeTokenTransfer account data.
  static deserializeAccount(data: Buffer): TrimmedAmount {
    const amount = data.readBigUInt64LE(0);
    const decimals = data.readUInt8(8);
    return new TrimmedAmount(amount, decimals);
  }

  static serialize(amount: TrimmedAmount): Buffer {
    const buffer = Buffer.alloc(9);
    buffer.writeUInt8(amount.decimals, 0);
    buffer.writeBigUInt64BE(amount.amount, 1);
    return buffer;
  }

  normalize(targetDecimals: number): bigint {
    if (this.decimals === targetDecimals) {
      return this.amount;
    } else if (this.decimals > targetDecimals) {
      return this.amount / BigInt(10 ** (this.decimals - targetDecimals));
    } else {
      return this.amount * BigInt(10 ** (targetDecimals - this.decimals));
    }
  }
}

export type SourceChainEvents =
  | "transfer-sent"
  | "send-transceiver-message"
  | "log-message-published";
export type TargetChainEvents =
  | "transfer-redeemed"
  | "received-relayed-message"
  | "message-attested-to";

export type NTTTransfer = {
  eventName: SourceChainEvents | TargetChainEvents;
  recipient?: string;
  amount?: bigint;
  fee?: bigint;
  recipientChain?: ChainId;
  emitterChain?: ChainId;
  sourceToken?: string;
  transceiverType?: "wormhole" | "axelar";
  digest: string;
};

export type EVMNTTManagerAttributes = TransactionFoundAttributes & {
  eventName: string;
  blockNumber: bigint;
  timestamp: number;
  txHash: string;
  gas: string;
  gasPrice: string;
  cost: bigint;
  nonce: string;
  gasUsed: string;
  effectiveGasPrice: string;
  recipient?: string;
  amount?: bigint;
  fee?: bigint;
  sourceToken?: string;
  digest: string;
};

export const getNttManagerMessageDigest = (
  emitterChain: number,
  message: NttManagerMessage<NativeTokenTransfer>
): string => {
  const chainIdBuffer = Buffer.alloc(2);
  chainIdBuffer.writeUInt16BE(emitterChain);
  const serialized = NttManagerMessage.serialize(message, NativeTokenTransfer.serialize);
  const digest = keccak256(Buffer.concat([chainIdBuffer, serialized]));
  return Buffer.from(digest).toString("hex");
};

export class WormholeTransceiverMessage<A> extends TransceiverMessage<A> {
  static prefix = Buffer.from([0x99, 0x45, 0xff, 0x10]);
}

type DecodedTransferSent = {
  recipient: string;
  refundAddr: string;
  amount: string;
  fee: string;
  recipientChain: number;
  msgSequence: number;
};

/// event TransferSent( bytes32 recipient, uint256 amount, uint256 fee, uint16 recipientChain, uint64 msgSequence);
export function decodeNttTransferSent(data: string): DecodedTransferSent {
  // There are 6 fields in this message.  All of them are 32 bytes (64 characters in hex)
  // If data starts with '0x', we need to remove it
  if (data.startsWith("0x")) {
    data = data.slice(2);
  }
  let retVal: DecodedTransferSent = {
    recipient: "",
    refundAddr: "",
    amount: "",
    fee: "",
    recipientChain: 0,
    msgSequence: 0,
  };
  if (data.length === 384) {
    retVal.recipient = data.slice(0, 64);
    retVal.refundAddr = "0x" + data.slice(64, 128);
    retVal.amount = "0x" + data.slice(128, 192);
    retVal.fee = "0x" + data.slice(192, 256);
    retVal.recipientChain = Number("0x" + data.slice(256, 320));
    retVal.msgSequence = Number("0x" + data.slice(320, 384));
  } else {
    throw new Error("Invalid data length.  Expected 384 characters.  Got " + data.length);
  }
  return retVal;
}
